package indexer

import (
	"errors"
	"github.com/spf13/viper"
	"net/http"
)

var (
	ErrUnknownIndexer       = errors.New("Unknown Indexer")
	DefaultDefinitionLoader DefinitionLoader
	Loader                  DefinitionLoader
)

func ListBuiltins() ([]string, error) {
	l := escLoader{http.Dir("")}
	return l.List()
}

func LoadEnabledDefinitions(conf interface{}) ([]*IndexerDefinition, error) {
	keys, err := DefaultDefinitionLoader.List()
	if err != nil {
		return nil, err
	}
	defs := []*IndexerDefinition{}
	for _, key := range keys {
		section := viper.Get(key)
		if section != nil {
			def, err := DefaultDefinitionLoader.Load(key)
			if err != nil {
				return nil, err
			}
			defs = append(defs, def)
		}
	}
	return defs, nil
}

type DefinitionLoader interface {
	//List - Lists available trackers.
	List() ([]string, error)
	//Load - Load a definition of an Indexer from it's name
	Load(key string) (*IndexerDefinition, error)
}

func init() {
	DefaultDefinitionLoader = defaultMultiLoader()
	//Start with the default loader.
	Loader = DefaultDefinitionLoader
}
