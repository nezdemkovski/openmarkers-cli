package models

type Profile struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	DateOfBirth  string  `json:"dateOfBirth"`
	Sex          string  `json:"sex"`
	IsPublic     bool    `json:"isPublic"`
	PublicHandle *string `json:"publicHandle"`
}

type ProfileData struct {
	User       ProfileUser    `json:"user"`
	Categories []CategoryData `json:"categories"`
}

type ProfileUser struct {
	ID           int     `json:"id,omitempty"`
	Name         string  `json:"name"`
	DateOfBirth  string  `json:"dateOfBirth,omitempty"`
	Sex          string  `json:"sex,omitempty"`
	PublicHandle *string `json:"publicHandle,omitempty"`
}

type CategoryData struct {
	ID         string         `json:"id"`
	Biomarkers []BiomarkerData `json:"biomarkers"`
}

type BiomarkerData struct {
	ID      string       `json:"id"`
	Type    string       `json:"type"`
	Unit    string       `json:"unit"`
	RefMin  *float64     `json:"refMin"`
	RefMax  *float64     `json:"refMax"`
	Results []ResultData `json:"results"`
}

type ResultData struct {
	ID    int    `json:"id,omitempty"`
	Date  string `json:"date"`
	Value any    `json:"value"`
}

type PublicProfile struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

type ImportCheck struct {
	Exists bool     `json:"exists"`
	User   *Profile `json:"user"`
}

type ImportResult struct {
	OK        bool `json:"ok"`
	ProfileID int  `json:"profile_id"`
}
