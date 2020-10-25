package indexer

import (
	"fmt"
	"strings"
)

func (s IndexerSelector) isAggregate() bool {
	return s.selector == "" || s.selector == "aggregate" || s.selector == "all" || strings.Contains(s.selector, ",")
}

type IndexerSelector struct {
	selector string
	parts    []string
}

//ResolveIndexId resolves the global aggregate index ID to a allowed index in the given scope
//if no id is given then the first index in the scope is used
func ResolveIndexId(scope Scope, id string) string {
	isGlobalAggregate := id == "" || id == "aggregate" || id == "all"
	if !isGlobalAggregate {
		return id
	}
	indexes := scope.Indexes()
	shouldChooseFirstIndex := id == ""
	//We're searching for the first index
	for ixId, ix := range indexes {
		if shouldChooseFirstIndex {
			return ixId
		}
		if _, ok := ix.(*Aggregate); ok {
			return ixId
		}
	}
	return id
}

func (s IndexerSelector) String() string {
	return fmt.Sprintf("%s:%s", s.selector, s.parts)
}

func (s IndexerSelector) shouldLoadAllIndexes() bool { //nolint:unused
	indexKeys := strings.Split(s.selector, ",")
	return s.isAggregate() && len(indexKeys) == 1
}

func (s IndexerSelector) Matches(name string) bool {
	if s.selector == "" {
		return true
	}
	if s.selector == "all" || s.selector == "aggregate" {
		return false
	}

	return contains(s.parts, name)
}

func (s IndexerSelector) Value() string {
	return s.selector
}

func newIndexerSelector(selector string) *IndexerSelector {
	return &IndexerSelector{selector: selector, parts: strings.Split(selector, ",")}
}
