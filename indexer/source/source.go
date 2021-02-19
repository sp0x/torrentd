package source

import (
	"io"
	"net/url"
)

type FetchOptions struct {
	URL            string
	Values         url.Values
	Method         string
	Encoding       string
	NoEncoding     bool
	ShouldDumpData bool
	FakeReferer    bool
}

func NewFetchOptions(url string) *FetchOptions {
	return &FetchOptions{
		URL: url,
	}
}

type FetchResult interface {
	ContentType() string
	Encoding() string
}

//go:generate mockgen -source source.go -destination=mocks/source.go -package=mocks
type ContentFetcher interface {
	Cleanup()
	Fetch(target *FetchOptions) (FetchResult, error)
	//Get(url string) error
	Post(url string, data url.Values, log bool) error
	URL() *url.URL
	Clone() ContentFetcher
	Open(options *FetchOptions) error
	Download(buffer io.Writer) (int64, error)
}
