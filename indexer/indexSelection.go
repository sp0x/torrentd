package indexer

import (
	"fmt"
	"strings"
)

const (
	indexSelectorAggregate = "aggregate"
	indexSelectorAll       = "all"
)

func (s Selector) isAggregate() bool {
	return s.selector == "" || s.selector == indexSelectorAggregate || s.selector == indexSelectorAll || strings.Contains(s.selector, ",")
}

// Selector that helps with index lookup
type Selector struct {
	selector string
	parts    []string
}

// ResolveIndexID resolves the global aggregate index ID to a allowed index in the given scope
// If no ID is given then the first aggregate is used. If an aggregate is not available, then the first working index is used.
func ResolveIndexID(scope Scope, id string) string {
	isGlobalAggregate := id == "" || id == indexSelectorAggregate || id == indexSelectorAll
	if !isGlobalAggregate {
		return id
	}
	indexes := scope.Indexes()
	shouldChooseFirstIndex := id == ""
	var firstWorkingIndex string
	// We're searching for the first index
	for ixID, indexes := range indexes {
		hasErrors := len(indexes.Errors()) > 0
		isAggregate := len(indexes) > 1
		if hasErrors && !isAggregate {
			continue
		}
		if firstWorkingIndex == "" {
			firstWorkingIndex = ixID
		}
		if isAggregate {
			return ixID
		}
	}
	if shouldChooseFirstIndex {
		return firstWorkingIndex
	}
	return id
}

func (s Selector) String() string {
	return fmt.Sprintf("%s:%s", s.selector, s.parts)
}

func (s Selector) shouldLoadAllIndexes() bool { //nolint:unused
	indexKeys := strings.Split(s.selector, ",")
	return s.isAggregate() && len(indexKeys) == 1
}

func (s Selector) Matches(name string) bool {
	if s.selector == "" || s.selector == indexSelectorAll {
		return true
	}
	if s.selector == indexSelectorAggregate {
		return false
	}

	return contains(s.parts, name)
}

func (s Selector) Value() string {
	return s.selector
}

func newIndexSelector(selector string) *Selector {
	parts := strings.Split(selector, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return &Selector{selector: selector, parts: parts}
}
