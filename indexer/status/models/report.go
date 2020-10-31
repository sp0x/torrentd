package models

type LatestResult struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Site        string `json:"site"`
	Link        string `json:"link"`
}

type IndexStatus struct {
	Index       string   `json:"index"`
	IsAggregate bool     `json:"is_aggregate"`
	Errors      []string `json:"errors"`
}
