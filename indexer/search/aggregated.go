package search

type AggregatedSearch struct {
	SearchContexts map[interface{}]Instance
	results        []ExternalResultItem
}

func (a *AggregatedSearch) GetResults() []ExternalResultItem {
	return a.results
}

func (a *AggregatedSearch) GetStartingIndex() int {
	panic("this is a stub")
}

func (s *AggregatedSearch) SetId(val string) {
	panic("this is a stub")
}

func (a *AggregatedSearch) SetStartIndex(key interface{}, i int) {
	srch := a.SearchContexts[key]
	if srch == nil {
		return
	}
	srch.SetStartIndex(nil, i)
}

func (a *AggregatedSearch) SetResults(results []ExternalResultItem) {
	a.results = results
}

func NewAggregatedSearch() *AggregatedSearch {
	ag := &AggregatedSearch{}
	ag.SearchContexts = make(map[interface{}]Instance)
	return ag
}
