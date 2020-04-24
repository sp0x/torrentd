package search

import "github.com/PuerkitoBio/goquery"

type SearchMode struct {
	Key             string
	Available       bool
	SupportedParams []string
}

type Search struct {
	DOM         *goquery.Selection
	Id          string
	CurrentPage int
	StartIndex  int
	Results     []ExternalResultItem
}

func (s *Search) GetDocument() *goquery.Selection {
	return s.DOM
}

type PaginationSearch struct {
	PageCount    uint
	StartingPage uint
}

type RunOptions struct {
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}
