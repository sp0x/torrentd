package indexer

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sort"
)

type MultipleDefinitionLoader []DefinitionLoader

func defaultMultiLoader() *MultipleDefinitionLoader {
	return &MultipleDefinitionLoader{
		defaultFsLoader(),
		embeddedLoader(),
		// escLoader{http.Dir("")},
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (ml MultipleDefinitionLoader) List(selector *IndexerSelector) ([]string, error) {
	allResults := map[string]struct{}{}

	for _, loader := range ml {
		result, err := loader.List(selector)
		if err != nil {
			return nil, err
		}
		for _, val := range result {
			allResults[val] = struct{}{}
		}
	}

	var results []string

	for key := range allResults {
		results = append(results, key)
	}

	sort.Strings(results)
	log.WithFields(log.Fields{"results": results, "loader": ml}).
		Debug("Multiple definitions loader listed indexes")
	return results, nil
}

func (ml MultipleDefinitionLoader) String() string {
	str := ""
	for ix, loader := range ml {
		if ix > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s", loader)
	}
	return "loaders[" + str + "]"
}

// Load an indexer with the matching name
func (ml MultipleDefinitionLoader) Load(key string) (*IndexerDefinition, error) {
	var def *IndexerDefinition
	// Go over each loader, until we reach the one that contains the definition for the indexer.
	for _, loader := range ml {
		if loader == nil {
			continue
		}
		loaded, err := loader.Load(key)
		if err != nil {
			log.Debugf("Couldn't load the Indexer `%s` using %s. Error : %s\n", key, loader, err)
			continue
		}
		// If it's newer than our last one
		if def == nil || loaded.Stats().ModTime.After(def.Stats().ModTime) { // If no definition is loaded so far, or the new one is newer
			def = loaded
		}
	}

	if def == nil {
		log.Infof("No loaders managed to load Indexer `%s` from any of these locations: \n", key)
		for _, ldr := range ml {
			log.Infof("%s\n", ldr)
		}
		return nil, ErrUnknownIndexer
	}

	return def, nil
}
