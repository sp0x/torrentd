package search

import (
	"errors"
	"fmt"
	"strings"
)

type Capability struct {
	Key             string
	Available       bool
	SupportedParams []string
}

type RangeField []string

type SearchStateIterator struct {
	// Stores the state of stateful search fields
	FieldState      map[string]*RangeFieldState
	ID              string
	StartingPage    uint
	CurrentPage     uint
	StopOnStale     bool
	PageCount       uint
	discoveredItems uint
	reachedStale    bool
}

type RunOptions struct {
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

type AggregatedSearch struct {
	SearchIterators map[interface{}]*SearchStateIterator
}

func NewIterator(query *Query) *SearchStateIterator {
	s := &SearchStateIterator{}
	s.PageCount = query.NumberOfPagesToFetch
	s.CurrentPage = query.Page
	s.StartingPage = query.Page

	s.FieldState = make(map[string]*RangeFieldState)
	s.StopOnStale = query.StopOnStale
	s.setFieldStateFromQuery(query)
	return s
}

func (s SearchStateIterator) GetItemsDiscoveredCount() uint {
	return s.discoveredItems
}

func (s SearchStateIterator) String() string {
	output := make([]string, len(s.FieldState))
	i := 0
	for fname, fval := range s.FieldState {
		val := fmt.Sprintf("{%s: %v}", fname, fval)
		output[i] = val
		i++
	}
	return strings.Join(output, ",")
}

func (s SearchStateIterator) hasNextFieldState() bool {
	for _, field := range s.FieldState {
		if field == nil {
			continue
		}
		if field.HasNext() {
			return true
		}
	}
	return false
}

func (s *SearchStateIterator) Next() (map[string]interface{}, uint) {
	fields := make(map[string]interface{})
	page := s.CurrentPage
	//If we have some fields, increment them
	for f, fstate := range s.FieldState {
		val := fstate.GetCurrent()
		fields[f] = val
		fstate.increment()
	}

	s.CurrentPage += 1
	return fields, page
}

func (s *SearchStateIterator) setFieldState(name string, value interface{}) {
	if value, ok := value.(RangeField); ok {
		s.FieldState[name] = &RangeFieldState{
			value[0],
			value[1],
			"",
		}
	}
}

func (s *SearchStateIterator) setFieldStateFromQuery(query *Query) {
	if query != nil {
		for fieldName, fieldValue := range query.Fields {
			s.setFieldState(fieldName, fieldValue)
		}
	}
}

func (s SearchStateIterator) GetFieldStateOrDefault(name string, args func() *RangeFieldState) (*RangeFieldState, interface{}) {
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

func (s *SearchStateIterator) UpdateIteratorState(r []ResultItemBase) {
	s.discoveredItems += uint(len(r))
	if !s.StopOnStale {
		return
	}
	for _, item := range r {
		if !item.IsNew() {
			s.reachedStale = true
			break
		}
	}
}

func (s SearchStateIterator) IsComplete() bool {
	hasNextFieldState := s.hasNextFieldState()
	isPageLimited := s.PageCount != 0
	pagesTraversed := s.CurrentPage - s.StartingPage
	hasExceededPages := isPageLimited &&
		(s.PageCount != 0 && s.PageCount <= pagesTraversed && pagesTraversed != 0)
	hasFields := s.FieldState != nil && len(s.FieldState) > 0

	if !hasFields && hasExceededPages {
		return true
	} else if isPageLimited && hasExceededPages {
		return true
	} else if hasFields && !hasNextFieldState {
		return true
	} else if s.StopOnStale && s.reachedStale {
		return true
	}
	return !hasNextFieldState && hasExceededPages
}

func NewRangeField(values ...string) RangeField {
	return values
}

func (a *AggregatedSearch) String() string {
	return "{aggregate}"
}

func NewAggregatedSearch() *AggregatedSearch {
	ag := &AggregatedSearch{}
	ag.SearchIterators = make(map[interface{}]*SearchStateIterator)
	return ag
}
