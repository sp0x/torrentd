package indexer

/**
This is part of https://github.com/cardigann/cardigann
*/

import (
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
	"io"
	"net/http"
)

type Info interface {
	GetId() string
	GetTitle() string
	GetLanguage() string
	GetLink() string
}

type Indexer interface {
	Info() Info
	Search(query torznab.Query, srch *search.Search) (*search.Search, error)
	Download(urlStr string) (io.ReadCloser, http.Header, error)
	Capabilities() torznab.Capabilities
	GetEncoding() string
	ProcessRequest(req *http.Request) (*http.Response, error)
}
