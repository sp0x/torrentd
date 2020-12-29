package search

import (
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type SearchMode struct {
	Key             string
	Available       bool
	SupportedParams []string
}

// An instance of a search
type Instance interface {
	GetStartingIndex() int
	GetResults() []ResultItemBase
	SetStartIndex(key interface{}, i int)
	SetResults(extracted []ResultItemBase)
	SetId(val string)
}

type RangeField []string

type Search struct {
	DOM         *goquery.Selection
	Id          string
	currentPage int
	StartIndex  int
	Results     []ResultItemBase
	FieldState  map[string]*rangeFieldState
}

func NewSearch(query *Query) *Search {
	s := &Search{}
	s.FieldState = make(map[string]*rangeFieldState)
	for fieldName, fieldValue := range query.Fields {
		switch value := fieldValue.(type) {
		case RangeField:
			s.FieldState[fieldName] = &rangeFieldState{
				value[0],
				value[1],
				"",
			}
		}
	}
	return s
}

type rangeFieldState struct {
	start   string
	end     string
	current string
}

func (r *rangeFieldState) Next() string {
	if r.current == "" {
		r.current = r.start
	} else if !r.HasNext() {
		return r.current
	} else {
		r.increment()
	}
	return r.current
}

func (r *rangeFieldState) HasNext() bool {
	return r.current != r.end
}

func (r *rangeFieldState) increment() {
	length := len(r.current)
	num, _ := strconv.Atoi(r.current)
	num += 1
	r.current = fmt.Sprintf("%0"+strconv.Itoa(length)+"d", num)
}

func (s *Search) GetStartingIndex() int {
	return s.StartIndex
}

func (s *Search) GetDocument() *goquery.Selection {
	return s.DOM
}

func (s *Search) SetStartIndex(_ interface{}, i int) {
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
