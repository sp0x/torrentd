package indexer

import (
	"errors"
	"github.com/spf13/viper"
)

var (
	ErrUnknownIndexer       = errors.New("unknown indexer")
	DefaultDefinitionLoader DefinitionLoader
	Loader                  DefinitionLoader
)

//ListBuiltInIndexes returns a list of all the embedded indexes that are supported.
func ListBuiltInIndexes() ([]string, error) {
	return DefaultDefinitionLoader.List()
}

//LoadEnabledDefinitions loads all of the definitions that are covered by the current definition loader (`Loader`).
func LoadEnabledDefinitions(conf interface{}) ([]*IndexerDefinition, error) {
	keys, err := Loader.List()
	if err != nil {
		return nil, err
	}
	var defs []*IndexerDefinition
	for _, key := range keys {
		section := viper.Get(key)
		if section != nil {
			def, err := Loader.Load(key)
			if err != nil {
				return nil, err
			}
			defs = append(defs, def)
		}
	}
	return defs, nil
}

//DefinitionLoader loads an index definition by name or lists the names of the supported indexes.
type DefinitionLoader interface {
	//List - Lists available trackers.
	List() ([]string, error)
	ListWithNames(names []string) ([]string, error)
	//Load - Load a definition of an Indexer from it's name
	Load(key string) (*IndexerDefinition, error)
}

func init() {
	DefaultDefinitionLoader = defaultMultiLoader()
	//Start with the default loader.
	Loader = DefaultDefinitionLoader
}
