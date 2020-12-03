package search

import (
	"github.com/PuerkitoBio/goquery"
)

type SearchMode struct {
	Key             string
	Available       bool
	SupportedParams []string
}

//An instance of a search
type Instance interface {
	GetStartingIndex() int
	GetResults() []ResultItemBase
	SetStartIndex(key interface{}, i int)
	SetResults(extracted []ResultItemBase)
	SetId(val string)
}

type Search struct {
	DOM         *goquery.Selection
	Id          string
	currentPage int
	StartIndex  int
	Results     []ResultItemBase
}

func (s *Search) GetStartingIndex() int {
	return s.StartIndex
}

func (s *Search) GetDocument() *goquery.Selection {
	return s.DOM
}

func (s *Search) SetStartIndex(key interface{}, i int) {
	s.StartIndex = i
}

func (s *Search) GetResults() []ResultItemBase {
	return s.Results
}

func (s *Search) SetResults(results []ResultItemBase) {
	s.Results = results
}

func (s *Search) SetId(val string) {
	s.Id = val
}

type PaginationSearch struct {
	PageCount    uint
	StartingPage uint
}

type RunOptions struct {
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}
