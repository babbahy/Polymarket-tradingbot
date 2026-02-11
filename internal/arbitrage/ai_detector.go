package arbitrage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/polymarket-arbitrage/internal/exchange"
	"go.uber.org/zap"
)

// AIDetector uses LLM to find market dependencies
type AIDetector struct {
	logger          *zap.Logger
	anthropicAPIKey string
	httpClient      *http.Client
	cache           map[string]*DependencyAnalysis
}

// DependencyAnalysis represents AI's analysis of market dependencies
type DependencyAnalysis struct {
	HasDependency     bool     `json:"has_dependency"`
	DependencyType    string   `json:"dependency_type"`
	Explanation       string   `json:"explanation"`
	ValidCombinations [][]int  `json:"valid_combinations"`
	Confidence        float64  `json:"confidence"`
	Constraint        string   `json:"constraint"`
}

func NewAIDetector(anthropicAPIKey string, logger *zap.Logger) *AIDetector {
	return &AIDetector{
		logger:          logger,
		anthropicAPIKey: anthropicAPIKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		cache: make(map[string]*DependencyAnalysis),
	}
}

// AnalyzeMarketPair uses Claude to detect dependencies between two markets
func (d *AIDetector) AnalyzeMarketPair(ctx context.Context, market1, market2 *exchange.Market) (*DependencyAnalysis, error) {
	// Check cache first
	cacheKey := market1.ID + ":" + market2.ID
	if cached, ok := d.cache[cacheKey]; ok {
		d.logger.Debug("Using cached analysis", zap.String("key", cacheKey))
		return cached, nil
	}

	prompt := d.buildAnalysisPrompt(market1, market2)
	
	d.logger.Info("🤖 AI analyzing market pair...",
		zap.String("market1", market1.Question),
		zap.String("market2", market2.Question),
	)

	analysis, err := d.callClaudeAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("Claude API failed: %w", err)
	}

	// Cache the result
	d.cache[cacheKey] = analysis

	d.logger.Info("🤖 AI analysis complete",
		zap.Bool("dependency_found", analysis.HasDependency),
		zap.String("type", analysis.DependencyType),
		zap.Float64("confidence", analysis.Confidence),
	)

	return analysis, nil
}

func (d *AIDetector) buildAnalysisPrompt(market1, market2 *exchange.Market) string {
	return fmt.Sprintf(`You are an expert at analyzing prediction markets for logical dependencies.

Analyze these two prediction markets:

MARKET A: "%s"
Outcomes: %v

MARKET B: "%s"
Outcomes: %v

Task: Identify if there's a logical dependency between these markets.

DEPENDENCY TYPES:
1. IMPLICATION: If one outcome is true, another MUST be true
   Example: "Republicans win PA by 5+" → "Trump wins PA"
   
2. MUTUAL_EXCLUSION: Two outcomes cannot both happen
   Example: "Biden wins" ⊕ "Trump wins"
   
3. CONDITIONAL: Complex probability relationships

Output ONLY this JSON (no markdown):
{
  "has_dependency": true or false,
  "dependency_type": "implication" | "mutual_exclusion" | "conditional" | "none",
  "explanation": "brief logical reasoning",
  "valid_combinations": [[0,0], [0,1], [1,0], [1,1]],
  "confidence": 0.0 to 1.0,
  "constraint": "P(A) <= P(B)"
}

valid_combinations encoding:
- [0,0] = A=NO, B=NO
- [0,1] = A=NO, B=YES
- [1,0] = A=YES, B=NO
- [1,1] = A=YES, B=YES

Exclude impossible combinations.`,
		market1.Question,
		getOutcomeNames(market1),
		market2.Question,
		getOutcomeNames(market2),
	)
}

func (d *AIDetector) callClaudeAPI(ctx context.Context, prompt string) (*DependencyAnalysis, error) {
	requestBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 2000,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.anthropic.com/v1/messages",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", d.anthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read body for error debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		d.logger.Error("Claude API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse JSON from Claude's response
	text := apiResponse.Content[0].Text
	text = extractJSON(text)

	var analysis DependencyAnalysis
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		d.logger.Error("Failed to parse Claude response",
			zap.String("response", text[:min(len(text), 500)]),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return &analysis, nil
}

// extractJSON extracts JSON from potentially markdown-wrapped text
func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	
	// Find first { and last }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	
	return text
}

func getOutcomeNames(market *exchange.Market) []string {
	names := make([]string, len(market.Tokens))
	for i, token := range market.Tokens {
		names[i] = token.Outcome
	}
	return names
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
