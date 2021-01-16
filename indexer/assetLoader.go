package indexer

import (
	"fmt"
	"path"
	"strings"

	"github.com/sp0x/torrentd/indexer/definitions"
)

// An indexer definition loader that uses embedded definitions.
type DefinitionDataLoader func(key string) ([]byte, error)

type AssetLoader struct {
	Names    []string
	Resolver DefinitionDataLoader
}

func embeddedLoader() DefinitionLoader {
	return getDefaultEmbeddedDefinitionSource()
}

func getDefaultEmbeddedDefinitionSource() DefinitionLoader {
	var src AssetLoader
	src.Names = definitions.GzipAssetNames()
	src.Resolver = func(key string) ([]byte, error) {
		fullname := fmt.Sprintf("definitions/%s.yml", key)
		data, err := definitions.GzipAsset(fullname)
		if err != nil {
			fullname = fmt.Sprintf("definitions/%s.yaml", key)
			data, err = definitions.GzipAsset(fullname)
			if err != nil {
				return nil, err
			}
		}
		data, _ = definitions.UnzipData(data)
		return data, nil
	}
	return &src
}

// CreateEmbeddedDefinitionSource creates a new definition loader from a set of resource names and a loader.
func CreateEmbeddedDefinitionSource(definitionNames []string, loader DefinitionDataLoader) DefinitionLoader {
	defLoader := &AssetLoader{}
	defLoader.Names = definitionNames
	defLoader.Resolver = loader
	return defLoader
}

// List all the names of the embedded definitions
func (l *AssetLoader) List(selector *IndexerSelector) ([]string, error) {
	results := make([]string, len(l.Names))
	for _, name := range l.Names {
		fname := path.Base(name)
		fname = strings.Replace(fname, ".yml", "", -1)
		fname = strings.Replace(fname, ".yaml", "", -1)
		if selector != nil && !selector.Matches(fname) {
			continue
		}
		results = append(results, fname)
	}
	return results, nil
}

func (l *AssetLoader) String() string {
	return "assets{}"
}

func (l *AssetLoader) ListWithNames(names []string) ([]string, error) {
	results := make([]string, len(l.Names))
	for _, name := range l.Names {
		if !contains(names, name) {
			continue
		}
		fname := path.Base(name)
		fname = strings.Replace(fname, ".yml", "", -1)
		fname = strings.Replace(fname, ".yaml", "", -1)
		results = append(results, fname)
	}
	return results, nil
}

// Load a definition with a given name
func (l *AssetLoader) Load(key string) (*IndexerDefinition, error) {
	data, err := l.Resolver(key)
	if err != nil {
		return nil, err
	}
	def, err := ParseDefinition(data)
	if err != nil {
		return def, err
	}
	def.stats.Source = "asset:" + key
	return def, err
}
