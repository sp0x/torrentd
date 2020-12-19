package source

import (
	"net/url"
)

type SearchTarget struct {
	Url    string
	Values url.Values
	Method string
}

func NewTarget(url string) *SearchTarget {
	return &SearchTarget{
		Url: url,
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
	FetchUrl(url string) error
	Post(url string, data url.Values, log bool) error
}
