package indexer

import (
	"errors"
	"fmt"
	"text/template"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/templates"
	"github.com/sp0x/torrentd/indexer/utils"
)

type SearchTemplateData struct {
	Query      *search.Query
	Keywords   string
	Categories []string
	Functions  template.FuncMap
	Search     *workerJob
}

func (s *SearchTemplateData) ApplyTo(name string, templateText string) (string, error) {
	return templates.ApplyTemplate(name, templateText, s)
}

func (s *SearchTemplateData) RangeValue(name string) string {
	fieldStates := s.Search.Fields
	if _, ok := fieldStates[name]; !ok {
		return ""
	}
	fieldState := fieldStates[name]
	// if !fieldState.HasNext() {
	// 	return ""
	// }
	// nextValue := fieldState.Next()
	return fieldState.(string)
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

func (s *SearchTemplateData) HasQueryField(name string) bool {
	if s.Query == nil {
		return false
	}
	_, ok := s.Query.Fields[name]
	return ok
}

func (s *SearchTemplateData) GetSearchFieldValue(fieldName string) (string, error) {
	if s.Search == nil || s.Search.Fields == nil {
		return "", errors.New("template has no search state")
	}
	fieldValue := s.Search.Fields[fieldName]
	if fieldValue == nil {
		return "", nil
	}
	switch value := fieldValue.(type) {
	default:
		return fmt.Sprint(value), nil
	}
}

func newSearchTemplateData(query *search.Query, srch *workerJob, localCategories []string) *SearchTemplateData {
	funcMap := utils.GetDefaultFunctionMap()
	searchData := &SearchTemplateData{
		query,
		query.Keywords(),
		localCategories,
		funcMap,
		srch,
	}
	return searchData
}
