package models

type Result struct {
	ID         int    `json:"id"`
	ProfileID  int    `json:"profile_id"`
	BiomarkerID string `json:"biomarker_id"`
	Date       string `json:"date"`
	Value      string `json:"value"`
	CreatedAt  string `json:"created_at"`
}

type BatchResult struct {
	Inserted int `json:"inserted"`
	Skipped  int `json:"skipped"`
}
