package search

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// An instance of a search
type Instance interface {
	GetStartingIndex() int
	GetResults() []ResultItemBase
	SetStartIndex(key interface{}, i int)
	SetResults(extracted []ResultItemBase)
	SetID(val string)
	HasFieldState() bool
	HasNext() bool
	IsComplete(q *Query) bool
	GetFieldState(name string, args func() *RangeFieldState) (*RangeFieldState, interface{})
}

type Capability struct {
	Key             string
	Available       bool
	SupportedParams []string
}

type RangeField []string

type Search struct {
	DOM        *goquery.Selection
	ID         string
	StartIndex int
	Results    []ResultItemBase
	// Stores the state of stateful search fields
	FieldState map[string]*RangeFieldState
}

func NewRangeField(values ...string) RangeField {
	return values
}

func NewSearch(query *Query) Instance {
	s := &Search{}
	s.FieldState = make(map[string]*RangeFieldState)
	s.setFieldStateFromQuery(query)
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

// IsDynamicSearch returns true if there are any fields with state
func (s *Search) HasFieldState() bool {
	for _, field := range s.FieldState {
		if field != nil {
			return true
		}
	}
	return false
}

func (s *Search) HasNext() bool {
	if !s.HasFieldState() {
		return false
	}
	for _, field := range s.FieldState {
		if field.HasNext() {
			return true
		}
	}
	return false
}

func (s *Search) IsComplete(q *Query) bool {
	hasNextPage := s.HasNext()
	hasExceededPages := q.PageCount <= q.Page && q.Page != 0 && q.PageCount != 0
	if hasExceededPages {
		return true
	}
	return !hasNextPage && hasExceededPages
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

func (s *Search) setFieldStateFromQuery(query *Query) {
	if query != nil {
		for fieldName, fieldValue := range query.Fields {
			s.setFieldState(fieldName, fieldValue)
		}
	}
}

func (s *Search) setFieldState(name string, value interface{}) {
	if value, ok := value.(RangeField); ok {
		s.FieldState[name] = &RangeFieldState{
			value[0],
			value[1],
			"",
		}
	}
}

func (s *Search) GetFieldState(name string, args func() *RangeFieldState) (*RangeFieldState, interface{}) {
	value, found := s.FieldState[name]
	if !found {
		if args == nil {
			return nil, errors.New("field has no state")
		}
		s.FieldState[name] = args()
		value = s.FieldState[name]
	}

	return value, nil
}

type PaginationSearch struct {
	PageCount    uint
	StartingPage uint
}

type RunOptions struct {
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

type AggregatedSearch struct {
	SearchContexts map[interface{}]Instance
	results        []ResultItemBase
}

func (a *AggregatedSearch) GetFieldState(name string, args func() *RangeFieldState) (*RangeFieldState, interface{}) {
	panic("this is a stub")
}

func (a *AggregatedSearch) HasFieldState() bool {
	panic("this is a stub")
}

func (a *AggregatedSearch) HasNext() bool {
	panic("this is a stub")
}

func (a *AggregatedSearch) IsComplete(q *Query) bool {
	for _, search := range a.SearchContexts {
		if !search.IsComplete(q) {
			return false
		}
	}

	return true
}

func (a *AggregatedSearch) GetResults() []ResultItemBase {
	return a.results
}

func (a *AggregatedSearch) GetStartingIndex() int {
	panic("this is a stub")
}

func (a *AggregatedSearch) SetID(val string) {
	panic("this is a stub")
}

func (a *AggregatedSearch) SetStartIndex(key interface{}, i int) {
	srch := a.SearchContexts[key]
	if srch == nil {
		return
	}
	srch.SetStartIndex(nil, i)
}

func (a *AggregatedSearch) SetResults(results []ResultItemBase) {
	a.results = results
}

func NewAggregatedSearch() *AggregatedSearch {
	ag := &AggregatedSearch{}
	ag.SearchContexts = make(map[interface{}]Instance)
	return ag
}
