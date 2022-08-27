package indexer

import (
	"errors"
	"fmt"
	"text/template"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/utils"
)

type SearchTemplateData struct {
	Query      *search.Query
	Keywords   string
	Categories []string
	Functions  template.FuncMap
	Search     *workerJob
}

func (s *SearchTemplateData) ApplyTo(fieldName string, templateText string) (string, error) {
	//fmap := s.Functions
	//fmap["rng"] = func(start, end string) string {
	//	return s.RangeValue(fieldName, start, end)
	//}
	return utils.ApplyTemplate(fieldName, templateText, s, s.Functions)
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
	if fieldValue==nil {
		return "", nil
	}
	switch value := fieldValue.(type) {
	default:
		return fmt.Sprint(value), nil
	}
}

func newSearchTemplateData(query *search.Query, srch *workerJob, localCategories []string) *SearchTemplateData {
	searchData := &SearchTemplateData{
		query,
		query.Keywords(),
		localCategories,
		utils.GetDefaultFunctionMap(),
		srch,
	}
	return searchData
}
