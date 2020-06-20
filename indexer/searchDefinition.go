package indexer

//searchBlock describes how search is done in an index.
type searchBlock struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
	//The number of results on each page, maximum.
	PageSize int `yaml:"pagesize"`
	//The maximum number of pages that we can fetch in a search.
	//1 if this is a single page search.
	MaxPages int             `yaml:"pages"`
	Inputs   inputsBlock     `yaml:"inputs,omitempty"`
	Rows     rowsBlock       `yaml:"rows"`
	Fields   fieldsListBlock `yaml:"fields"`
	Context  fieldsListBlock `yaml:"context"`
	//Key for indexing the results
	Key stringorslice `yaml:"key"`
}

//IsSinglePage figure out if the search is for a single page.
func (sb *searchBlock) IsSinglePage() bool {
	if sb.MaxPages == 1 || sb.Inputs == nil {
		return true
	}
	return false
}
