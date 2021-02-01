package source

import (
	"net/url"
)

type SearchTarget struct {
	URL    string
	Values url.Values
	Method string
}

func NewTarget(url string) *SearchTarget {
	return &SearchTarget{
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
	Fetch(target *SearchTarget) (FetchResult, error)
	FetchURL(url string) error
	Post(url string, data url.Values, log bool) error
	URL() *url.URL
}
