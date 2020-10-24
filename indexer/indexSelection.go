package indexer

import "strings"

func (s IndexerSelector) isAggregate() bool {
	return s.selector == "" || s.selector == "aggregate" || s.selector == "all" || strings.Contains(s.selector, ",")
}

type IndexerSelector struct {
	selector string
	parts    []string
}

//ResolveIndexId resolves the global aggregate index ID to a allowed index in the given scope
func ResolveIndexId(scope Scope, id string) string {
	isGlobalAggregate := id == "" || id == "aggregate" || id == "all"
	if !isGlobalAggregate {
		return id
	}
	indexes := scope.Indexes()
	for ixId, ix := range indexes {
		if _, ok := ix.(*Aggregate); ok {
			return ixId
		}
	}
	return id
}

func (s IndexerSelector) shouldLoadAllIndexes() bool {
	indexKeys := strings.Split(s.selector, ",")
	return s.isAggregate() && len(indexKeys) == 1
}

func (s IndexerSelector) Matches(name string) bool {
	if s.selector == "" || s.selector == "all" || s.selector == "aggregate" {
		return false
	}
	return contains(s.parts, name)
}

func (s IndexerSelector) Value() string {
	return s.selector
}

func newIndexerSelector(selector string) IndexerSelector {
	return IndexerSelector{selector: selector, parts: strings.Split(selector, ",")}
}
