package indexer

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
)

type multiLoader []DefinitionLoader

func defaultMultiLoader() *multiLoader {
	return &multiLoader{
		newFsLoader(),
		escLoader{http.Dir("")},
	}
}

func (ml multiLoader) List() ([]string, error) {
	allResults := map[string]struct{}{}

	for _, loader := range ml {
		result, err := loader.List()
		if err != nil {
			return nil, err
		}
		for _, val := range result {
			allResults[val] = struct{}{}
		}
	}

	results := []string{}

	for key := range allResults {
		results = append(results, key)
	}

	sort.Sort(sort.StringSlice(results))
	return results, nil
}

//
func (ml multiLoader) Load(key string) (*IndexerDefinition, error) {
	var def *IndexerDefinition

	for _, loader := range ml {
		if loader == nil {
			continue
		}
		loaded, err := loader.Load(key)
		if err != nil {
			continue
		}
		if def == nil || loaded.Stats().ModTime.After(def.Stats().ModTime) { // If no definition is loaded so far, or the new one is newer
			def = loaded
		}
	}

	if def == nil {
		log.Infof("Couldn't load the indexer `%s` from any of these locations: \n", key)
		for _, ldr := range ml {
			log.Infof("%s\n", ldr)
		}
		return nil, ErrUnknownIndexer
	}

	return def, nil
}
