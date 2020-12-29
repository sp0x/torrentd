package indexer

import (
	"io"
	"net/http"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
)

type Info interface {
	GetId() string
	GetTitle() string
	GetLanguage() string
	GetLink() string
}

type ResponseProxy struct {
	Reader            io.ReadCloser
	ContentLengthChan chan int64
}

//go:generate mockgen -source indexer.go -destination=mocks/indexer.go -package=mocks
type Indexer interface {
	Info() Info
	GetDefinition() *IndexerDefinition
	Search(query *search.Query, srch search.Instance) (search.Instance, error)
	Download(urlStr string) (*ResponseProxy, error)
	Capabilities() torznab.Capabilities
	GetEncoding() string
	ProcessRequest(req *http.Request) (*http.Response, error)
	Open(s search.ResultItemBase) (*ResponseProxy, error)
	// HealthCheck if the Indexer works.
	// This might be needed to validate the search result extraction.
	HealthCheck() error
	// The maximum number of pages we can search
	MaxSearchPages() uint
	SearchIsSinglePaged() bool
	Errors() []string
	GetStorage() storage.ItemStorage
	Site() string
}
