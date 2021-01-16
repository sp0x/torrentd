package indexer

type State struct {
	values map[string]interface{}
}

func (is *State) Set(key string, val interface{}) {
	is.values[key] = val
}

func (is *State) GetBool(key string) bool {
	if _, ok := is.values[key]; !ok {
		return false
	}
	return is.values[key].(bool)
}

func (is *State) Has(key string) bool {
	if _, ok := is.values[key]; ok {
		return true
	}
	return false
}

func defaultIndexerState() *State {
	is := &State{}
	is.values = make(map[string]interface{})
	is.Set("loggedIn", false)
	return is
}
