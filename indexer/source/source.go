package source

import "net/url"

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

type ContentFetcher interface {
	Cleanup()
	Fetch(target *SearchTarget) error
	FetchUrl(url string) error
	Post(url string, data url.Values, log bool) error
}
