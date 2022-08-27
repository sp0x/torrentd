package indexer

import (
	"errors"

	"github.com/spf13/viper"
)

var (
	ErrUnknownIndexer       = errors.New("unknown indexer")
	DefaultDefinitionLoader DefinitionLoader
	loader                  DefinitionLoader
)

// ListBuiltInIndexes returns a list of all the embedded indexMap that are supported.
func ListBuiltInIndexes() ([]string, error) {
	return DefaultDefinitionLoader.ListAvailableIndexes(nil)
}

// LoadEnabledDefinitions loads all the definitions that are covered by the current definition loader (`Loader`).
func LoadEnabledDefinitions(conf interface{}) ([]*Definition, error) {
	keys, err := GetIndexDefinitionLoader().ListAvailableIndexes(nil)
	if err != nil {
		return nil, err
	}
	var defs []*Definition
	for _, key := range keys {
		section := viper.Get(key)
		if section != nil {
			def, err := GetIndexDefinitionLoader().Load(key)
			if err != nil {
				return nil, err
			}
			defs = append(defs, def)
		}
	}
	return defs, nil
}

// DefinitionLoader loads an index definition by name or lists the names of the supported indexMap.
type DefinitionLoader interface {
	// ListAvailableIndexes - Lists available indexes
	ListAvailableIndexes(selector *Selector) ([]string, error)
	// Load - Load a definition of an Indexer from the name of the index
	Load(key string) (*Definition, error)
}

func init() {
	DefaultDefinitionLoader = defaultMultiLoader()
	// Start with the default loader.
	loader = DefaultDefinitionLoader
}

func GetIndexDefinitionLoader() DefinitionLoader {
	return loader
}