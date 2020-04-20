package torznab

/**
This is part of https://github.com/cardigann/cardigann
*/

import (
	"github.com/sp0x/rutracker-rss/indexer/search"
	"io"
	"net/http"
)

type Info struct {
	ID          string
	Title       string
	Description string
	Link        string
	Language    string
	Category    string
}

type Indexer interface {
	Info() Info
	Search(query Query) ([]search.ResultItem, error)
	Download(urlStr string) (io.ReadCloser, http.Header, error)
	Capabilities() Capabilities
}
