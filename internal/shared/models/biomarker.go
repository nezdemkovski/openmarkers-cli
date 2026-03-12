package models

type Biomarker struct {
	ID         string   `json:"id"`
	CategoryID string   `json:"category_id"`
	Unit       string   `json:"unit"`
	RefMin     *float64 `json:"ref_min"`
	RefMax     *float64 `json:"ref_max"`
	Type       string   `json:"type"`
}
