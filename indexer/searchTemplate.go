package indexer

import (
	"fmt"

	"github.com/sp0x/torrentd/indexer/templates"

	"github.com/sp0x/torrentd/indexer/search"
)

type SearchTemplateData struct {
	Query      *search.Query
	Keywords   string
	Categories []string
	Context    RunContext
}

func (s *SearchTemplateData) ApplyTo(name string, templateText string) (string, error) {
	return templates.ApplyTemplate(name, templateText, s)
}

func (s *SearchTemplateData) HasQueryField(name string) bool {
	_, ok := s.Query.Fields[name]
	return ok
}

func (s *SearchTemplateData) ApplyField(name string) string {
	fieldValue := s.Query.Fields[name]
	switch value := fieldValue.(type) {
	case search.RangeField:
		return s.RangeValue(name)
	default:
		return fmt.Sprint(value)
	}
}

func (s *SearchTemplateData) RangeValue(name string) string {
	fieldStates := s.Context.Search.FieldState
	if _, ok := fieldStates[name]; !ok {
		return ""
	}
	fieldState := fieldStates[name]
	if !fieldState.HasNext() {
		return ""
	}
	nextValue := fieldState.Next()
	return nextValue
}

func newSearchTemplateData(query *search.Query, localCategories []string, context RunContext) *SearchTemplateData {
	searchData := &SearchTemplateData{
		query,
		query.Keywords(),
		localCategories,
		context,
	}
	return searchData
}
