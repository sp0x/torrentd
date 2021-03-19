package indexer

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
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

func (ml MultipleDefinitionLoader) List(selector *Selector) ([]string, error) {
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

	results := make([]string, len(allResults))

	i := 0
	for key := range allResults {
		results[i] = key
		i++
	}

	sort.Strings(results)
	log.WithFields(log.Fields{"results": results, "loader": ml}).
		Debug("Multiple definitions loader listed indexesCollection")
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
func (ml MultipleDefinitionLoader) Load(key string) (*Definition, error) {
	var def *Definition
	// Go over each loader, until we reach the one that contains the definition for the indexer.
	for _, loader := range ml {
		if loader == nil {
			continue
		}
		loaded, err := loader.Load(key)
		if err != nil {
			log.Debugf("Couldn't load the Index `%s` using %s. Error : %s\n", key, loader, err)
			continue
		}
		// If it's newer than our last one
		if def == nil || loaded.Stats().ModTime.After(def.Stats().ModTime) { // If no definition is loaded so far, or the new one is newer
			def = loaded
		}
	}

	if def == nil {
		log.Infof("No loaders managed to load Index `%s` from any of these locations: \n", key)
		for _, ldr := range ml {
			log.Infof("%s\n", ldr)
		}
		return nil, ErrUnknownIndexer
	}

	return def, nil
}
