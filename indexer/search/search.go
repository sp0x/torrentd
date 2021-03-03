package search

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Mode struct {
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
	SetID(val string)
	IsDynamicSearch() bool
	HasCompletedDynamicSearch() bool
}

type RangeField []string

type Search struct {
	DOM        *goquery.Selection
	ID         string
	StartIndex int
	Results    []ResultItemBase
	// Stores the state of stateful search fields
	FieldState map[string]*rangeFieldState
}

func NewSearch(query *Query) Instance {
	s := &Search{}
	s.FieldState = make(map[string]*rangeFieldState)
	for fieldName, fieldValue := range query.Fields {
		if value, ok := fieldValue.(RangeField); ok {
			s.FieldState[fieldName] = &rangeFieldState{
				value[0],
				value[1],
				"",
			}
		}
	}
	return s
}

func (s *Search) String() string {
	output := make([]string, len(s.FieldState))
	i := 0
	for fname, fval := range s.FieldState {
		val := fmt.Sprintf("{%s: %v}", fname, fval)
		output[i] = val
		i++
	}
	return strings.Join(output, ",")
}

func (s *Search) IsDynamicSearch() bool {
	for _, field := range s.FieldState {
		if field != nil {
			return true
		}
	}
	return false
}

func (s *Search) HasCompletedDynamicSearch() bool {
	for _, field := range s.FieldState {
		if field.HasNext() {
			return false
		}
	}
	return true
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

func (s *Search) SetID(val string) {
	s.ID = val
}

type PaginationSearch struct {
	PageCount    uint
	StartingPage uint
}

type RunOptions struct {
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}
