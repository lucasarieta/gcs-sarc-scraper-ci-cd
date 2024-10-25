package model

type Class struct {
	Url       string                   `json:"url"`
	ClassCode string                   `json:"classCode"`
	Credits   int                      `json:"credits"`
	Title     string                   `json:"title"`
	Group     string                   `json:"group"`
	Building  string                   `json:"building"`
	Room      string                   `json:"room"`
	Year      int                      `json:"year"`
	Semester  int                      `json:"semester"`
	Sessions  []map[string]interface{} `json:"sessions"`
}
