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

// ListBuiltInIndexes returns a list of all the embedded indexesCollection that are supported.
func ListBuiltInIndexes() ([]string, error) {
	return DefaultDefinitionLoader.List(nil)
}

// LoadEnabledDefinitions loads all of the definitions that are covered by the current definition loader (`Loader`).
func LoadEnabledDefinitions(conf interface{}) ([]*Definition, error) {
	keys, err := Loader.List(nil)
	if err != nil {
		return nil, err
	}
	var defs []*Definition
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

// DefinitionLoader loads an index definition by name or lists the names of the supported indexesCollection.
type DefinitionLoader interface {
	// List - Lists available trackers.
	List(selector *Selector) ([]string, error)
	// Load - Load a definition of an Indexer from it's name
	Load(key string) (*Definition, error)
}

func init() {
	DefaultDefinitionLoader = defaultMultiLoader()
	// Start with the default loader.
	Loader = DefaultDefinitionLoader
}
