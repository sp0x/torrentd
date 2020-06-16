package indexer

import (
	"fmt"
	"github.com/sp0x/torrentd/indexer/definitions"
	"path"
	"strings"
)

//An indexer definition loader that uses embedded definitions.
type DefinitionDataLoader func(key string) ([]byte, error)

type EmbeddedDefinitionSource interface {
	GetNames() []string
	GetData(key string) ([]byte, error)
}

type AssetLoader struct {
	Names    []string
	Resolver DefinitionDataLoader
}

func embeddedLoader() DefinitionLoader {
	return GetDefaultEmbeddedDefinitionSource()
}

func GetDefaultEmbeddedDefinitionSource() DefinitionLoader {
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

func CreateEmbeddedDefinitionSource(definitionNames []string, loader DefinitionDataLoader) DefinitionLoader {
	defLoader := &AssetLoader{}
	defLoader.Names = definitionNames
	defLoader.Resolver = loader
	return defLoader
}

func (l *AssetLoader) GetNames() []string {
	return l.Names
}

func (l *AssetLoader) GetData(key string) ([]byte, error) {
	return l.Resolver(key)
}

//List all the names of the embedded definitions
func (l *AssetLoader) List() ([]string, error) {
	names := definitions.GzipAssetNames()
	var results []string
	for _, name := range names {
		fname := path.Base(name)
		fname = strings.Replace(fname, ".yml", "", -1)
		fname = strings.Replace(fname, ".yaml", "", -1)
		results = append(results, fname)
	}
	return results, nil
}

//Load a definition with a given name
func (l *AssetLoader) Load(key string) (*IndexerDefinition, error) {
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
	def, err := ParseDefinition(data)
	if err != nil {
		return def, err
	}
	def.stats.Source = "asset:" + fullname
	return def, err
}
