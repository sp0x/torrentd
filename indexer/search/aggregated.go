package search

type AggregatedSearch struct {
	SearchContexts map[interface{}]Instance
	results        []ResultItemBase
}

func (a *AggregatedSearch) HasFieldState() bool {
	panic("implement me")
}

func (a *AggregatedSearch) HasNext() bool {
	panic("this is a stub")
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
