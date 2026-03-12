package presentation

type ErrMsg struct{ Err error }
type ProfilesMsg struct{ Profiles []ProfileItem }
type ResultsMsg struct{ Results []ResultItem }

type ProfileItem struct {
	ID   int
	Name string
	DOB  string
	Sex  string
}

type ResultItem struct {
	ID          int
	BiomarkerID string
	Date        string
	Value       string
}
