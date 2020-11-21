package stats

import "time"

type Stats struct {
	Namespaces []NamespaceStats `json:"namespaces"`
}

func (s *Stats) GetNamespace(name string) *NamespaceStats {
	for i, nsStat := range s.Namespaces {
		if name == nsStat.Name {
			return &s.Namespaces[i]
		}
	}
	return nil
}

type NamespaceStats struct {
	Name        string    `json:"name"`
	RecordCount int       `json:"record_count"`
	LastUpdated time.Time `json:"last_updated"`
}
