package search

import "github.com/PuerkitoBio/goquery"

type Search struct {
	DOM         *goquery.Selection
	Id          string
	CurrentPage int
	StartIndex  int
	Results     []ResultItem
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
