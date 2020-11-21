package stats

import "time"

type Stats struct {
	Namespaces []NamespaceStats `json:"namespaces"`
}

type NamespaceStats struct {
	Name        string    `json:"name"`
	RecordCount int       `json:"record_count"`
	LastUpdated time.Time `json:"last_updated"`
}
