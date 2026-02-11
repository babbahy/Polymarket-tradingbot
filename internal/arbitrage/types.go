package arbitrage

type OpportunityType string

const (
	OpportunityTypeBinary       OpportunityType = "binary"
	OpportunityTypeCategorical  OpportunityType = "categorical"
	OpportunityTypeCombinatorial OpportunityType = "combinatorial"
)

type Opportunity struct {
	Type        OpportunityType
	MarketID    string
	MarketName  string
	ProfitBps   int
	Timestamp   int64
	Prices      []float64
	Strategy    string
}
