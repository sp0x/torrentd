package indexer

//searchBlock describes how search is done in an index.
type searchBlock struct {
	Path     string          `yaml:"path"`
	Method   string          `yaml:"method"`
	PageSize int             `yaml:"pagesize"`
	Inputs   inputsBlock     `yaml:"inputs,omitempty"`
	Rows     rowsBlock       `yaml:"rows"`
	Fields   fieldsListBlock `yaml:"fields"`
	Context  fieldsListBlock `yaml:"context"`
	//Key for indexing the results
	Key stringorslice `yaml:"key"`
}
