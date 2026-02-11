package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Client struct {
	gammaURL    string
	clobURL     string
	apiKey      string
	apiSecret   string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	logger      *zap.Logger
}

func NewClient(baseURL, apiKey, apiSecret string, rps int, logger *zap.Logger) *Client {
	return &Client{
		gammaURL:    "https://gamma-api.polymarket.com",
		clobURL:     baseURL,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		rateLimiter: rate.NewLimiter(rate.Limit(rps), 1),
		logger:      logger,
	}
}

func (c *Client) GetMarkets(ctx context.Context) ([]Market, error) {
	allMarkets := make([]Market, 0)
	
	// Fetch multiple pages to get more markets
	for offset := 0; offset < 500; offset += 100 {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit wait: %w", err)
		}

		url := fmt.Sprintf("%s/markets?closed=false&limit=100&offset=%d", c.gammaURL, offset)
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		var gammaMarkets []GammaMarket
		if err := json.Unmarshal(body, &gammaMarkets); err != nil {
			c.logger.Error("Failed to parse Gamma response", 
				zap.Error(err),
				zap.Int("offset", offset),
			)
			break
		}

		// If we got less than 100, we've reached the end
		if len(gammaMarkets) == 0 {
			break
		}

		converted := c.convertGammaMarkets(gammaMarkets)
		allMarkets = append(allMarkets, converted...)
		
		c.logger.Debug("Fetched page", 
			zap.Int("offset", offset),
			zap.Int("fetched", len(gammaMarkets)),
			zap.Int("valid", len(converted)),
		)
		
		// If less than 100, we're at the end
		if len(gammaMarkets) < 100 {
			break
		}
	}

	c.logger.Info("All markets fetched", zap.Int("total", len(allMarkets)))
	
	return allMarkets, nil
}

func (c *Client) convertGammaMarkets(gammaMarkets []GammaMarket) []Market {
	markets := make([]Market, 0, len(gammaMarkets))
	
	for _, gm := range gammaMarkets {
		// Parse outcome prices - they come as JSON array of STRINGS
		var priceStrings []string
		if err := json.Unmarshal([]byte(gm.OutcomePrices), &priceStrings); err != nil {
			continue
		}
		
		// Convert string prices to floats
		prices := make([]float64, len(priceStrings))
		for i, ps := range priceStrings {
			price, err := strconv.ParseFloat(ps, 64)
			if err != nil {
				continue
			}
			prices[i] = price
		}
		
		// Parse outcomes
		var outcomes []string
		if err := json.Unmarshal([]byte(gm.Outcomes), &outcomes); err != nil {
			continue
		}
		
		// Sanity check
		if len(outcomes) != len(prices) {
			continue
		}
		
		// Create tokens
		tokens := make([]Token, len(outcomes))
		for i := range outcomes {
			tokens[i] = Token{
				Outcome: outcomes[i],
				Price:   prices[i],
			}
		}
		
		markets = append(markets, Market{
			ID:          gm.ConditionID,
			Question:    gm.Question,
			Active:      gm.Active,
			Closed:      gm.Closed,
			Tokens:      tokens,
			MarketSlug:  gm.Slug,
			ConditionID: gm.ConditionID,
		})
	}
	
	return markets
}

func (c *Client) Close() error {
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
