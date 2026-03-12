package models

type SnapshotEntry struct {
	CategoryID  string   `json:"categoryId"`
	BiomarkerID string   `json:"biomarkerId"`
	Unit        string   `json:"unit"`
	RefMin      *float64 `json:"refMin"`
	RefMax      *float64 `json:"refMax"`
	Value       NumValue `json:"value"`
	OutOfRange  bool     `json:"outOfRange"`
}

type TrendEntry struct {
	BiomarkerID   string   `json:"biomarkerId"`
	CategoryID    string   `json:"categoryId"`
	Direction     string   `json:"direction"`
	RateChange    float64  `json:"rateChange"`
	OverallChange float64  `json:"overallChange"`
	TrendWarning  bool     `json:"trendWarning"`
	Improving     *bool    `json:"improving"`
	LatestValue   NumValue `json:"latestValue"`
	LatestDate    string   `json:"latestDate"`
}

type CompareEntry struct {
	CategoryID  string   `json:"categoryId"`
	BiomarkerID string   `json:"biomarkerId"`
	Unit        string   `json:"unit"`
	RefMin      *float64 `json:"refMin"`
	RefMax      *float64 `json:"refMax"`
	V1          NumValue `json:"v1"`
	V2          NumValue `json:"v2"`
	Delta       NumValue `json:"delta"`
	DeltaPct    NumValue `json:"deltaPct"`
	Out1        bool     `json:"out1"`
	Out2        bool     `json:"out2"`
}

type CorrelationGroup struct {
	ID      string   `json:"id"`
	Matched []string `json:"matched"`
}

type BioAgeEntry struct {
	Date           string        `json:"date"`
	PhenoAge       float64       `json:"phenoAge"`
	ChronoAge      float64       `json:"chronoAge"`
	Delta          float64       `json:"delta"`
	MortalityScore float64       `json:"mortalityScore"`
	DnamAge        float64       `json:"dnamAge"`
	DnamMortality  float64       `json:"dnamMortality"`
	Xb             float64       `json:"xb"`
	Scores         []BioAgeScore `json:"scores"`
}

type BioAgeScore struct {
	ID     string   `json:"id"`
	Value  NumValue `json:"value"`
	Unit   string   `json:"unit"`
	Score  float64  `json:"score"`
	Date   string   `json:"date"`
	RefMin *float64 `json:"refMin"`
	RefMax *float64 `json:"refMax"`
}

type AnalysisPrompt struct {
	Prompt string `json:"prompt"`
}

type DaysSinceEntry struct {
	CategoryID string  `json:"categoryId"`
	Days       *int    `json:"days"`
	LastDate   *string `json:"lastDate"`
}
