package indexer

import (
	"fmt"
	"github.com/sp0x/torrentd/indexer/definitions"
	"path"
	"strings"
)

type AssetLoader struct{}

func embeddedLoader() DefinitionLoader {
	return &AssetLoader{}
}

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
