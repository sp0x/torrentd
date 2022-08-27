package indexer

import (
	"io"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
)

type Info interface {
	GetID() string
	GetTitle() string
	GetLanguage() string
	GetLink() string
}

type ResponseProxy struct {
	Reader            io.ReadCloser
	ContentLengthChan chan int64
}

func NewResponseProxy() (*ResponseProxy, *io.PipeWriter) {
	pipeR, pipeW := io.Pipe()
	return &ResponseProxy{
		Reader:            pipeR,
		ContentLengthChan: make(chan int64),
	}, pipeW
}

//go:generate mockgen -source indexer.go -destination=indexerMocks.go -package=indexer
type Indexer interface {
	Info() Info
	GetDefinition() *Definition
	Search(query *search.Query, srch *workerJob) ([]search.ResultItemBase, error)
	Download(urlStr string) (*ResponseProxy, error)
	Capabilities() torznab.Capabilities
	GetEncoding() string
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
