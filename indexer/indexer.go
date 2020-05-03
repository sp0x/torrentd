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
	Search(query torznab.Query, srch search.Instance) (search.Instance, error)
	Download(urlStr string) (io.ReadCloser, error)
	Capabilities() torznab.Capabilities
	GetEncoding() string
	ProcessRequest(req *http.Request) (*http.Response, error)
	Open(s *search.ExternalResultItem) (io.ReadCloser, error)
	//Check if the Indexer works ok.
	//This might be needed to validate the search result extraction.
	Check() error
}
