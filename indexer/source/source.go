package source

import (
	"io"
	"net/http"
	"net/url"
)

type FetchOptions struct {
	ShouldDumpData bool
	FakeReferer    bool
}

type RequestOptions struct {
	URL        string
	Values     url.Values
	Method     string
	Encoding   string
	NoEncoding bool
	CookieJar  http.CookieJar
	Referer    *url.URL
}

func NewRequestOptions(url string) *RequestOptions {
	return &RequestOptions{
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
	Fetch(target *RequestOptions) (FetchResult, error)
	//Get(url string) error
	Post(options *RequestOptions) error
	URL() *url.URL
	Clone() ContentFetcher
	Open(options *RequestOptions) error
	Download(buffer io.Writer) (int64, error)
}
