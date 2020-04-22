package indexer

type IndexerState struct {
	values map[string]interface{}
}

func (is *IndexerState) Set(key string, val interface{}) {
	is.values[key] = val
}

func (is *IndexerState) GetBool(key string) bool {
	if _, ok := is.values[key]; !ok {
		return false
	}
	return is.values[key].(bool)
}

func (is *IndexerState) Has(key string) bool {
	if _, ok := is.values[key]; ok {
		return true
	}
	return false
}

func defaultIndexerState() *IndexerState {
	is := &IndexerState{}
	is.values = make(map[string]interface{})
	is.Set("loggedIn", false)
	return is
}
