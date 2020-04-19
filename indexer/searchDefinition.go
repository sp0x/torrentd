package indexer

type searchBlock struct {
	Path   string          `yaml:"path"`
	Method string          `yaml:"method"`
	Inputs inputsBlock     `yaml:"inputs,omitempty"`
	Rows   rowsBlock       `yaml:"rows"`
	Fields fieldsListBlock `yaml:"fields"`
}
