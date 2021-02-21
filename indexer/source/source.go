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
	URL        *url.URL
	Values     url.Values
	Method     string
	Encoding   string
	NoEncoding bool
	CookieJar  http.CookieJar
	Referer    *url.URL
}

func NewRequestOptions(destURL *url.URL) *RequestOptions {
	return &RequestOptions{
		URL: destURL,
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
	Post(options *RequestOptions) error
	URL() *url.URL
	Clone() ContentFetcher
	Open(options *RequestOptions) (FetchResult, error)
	Download(buffer io.Writer) (int64, error)
	SetErrorHandler(callback func(options *RequestOptions))
}
