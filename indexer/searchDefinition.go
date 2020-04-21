package indexer

type searchBlock struct {
	Path     string          `yaml:"path"`
	Method   string          `yaml:"method"`
	PageSize int             `yaml:"pagesize"`
	Inputs   inputsBlock     `yaml:"inputs,omitempty"`
	Rows     rowsBlock       `yaml:"rows"`
	Fields   fieldsListBlock `yaml:"fields"`
	Context  fieldsListBlock `yaml:"context"`
}
