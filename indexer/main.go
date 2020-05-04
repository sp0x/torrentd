package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/sp0x/rutracker-rss/indexer/categories"
)

var indexers map[string]Indexer

func init() {
	indexers = make(map[string]Indexer)
}

//Lookup finds the matching Indexer.
func Lookup(config config.Config, key string) (Indexer, error) {
	if _, ok := indexers[key]; !ok {
		var indexer Indexer
		var err error
		if key == "aggregate" || key == "all" {
			indexer, err = CreateAggregate(config)
		} else {
			indexer, err = CreateIndexer(config, key)
		}
		if err != nil {
			return nil, err
		}
		indexers[key] = indexer
	}
	return indexers[key], nil
}

func CreateAggregateForCategories(config config.Config, cats categories.Categories) (Indexer, error) {
	ixrKeys, err := DefaultDefinitionLoader.List()
	if err != nil {
		return nil, err
	}
	var indexers []Indexer
	for _, key := range ixrKeys {
		ifaceConfig, _ := config.GetSite(key) //Get all the configured indexers
		ixr, err := Lookup(config, key)
		if !ixr.Capabilities().HasCategories(cats) {
			continue
		}
		indexers = append(indexers, ixr)
	}

}

//CreateAggregate gets you an aggregate of all the valid configured indexers
//this includes indexers that don't need a login.
func CreateAggregate(config config.Config) (Indexer, error) {
	keys, err := DefaultDefinitionLoader.List()
	if err != nil {
		return nil, err
	}
	var indexers []Indexer
	for _, key := range keys {
		//Get the site configuration
		ifaceConfig, _ := config.GetSite(key) //Get all the configured indexers
		if ifaceConfig != nil && len(ifaceConfig) > 0 {
			indexer, err := Lookup(config, key)
			if err != nil {
				return nil, err
			}
			indexers = append(indexers, indexer)
		} else {
			//Indexer might not be configured
			//indexer, err := Lookup(config, key)
			//if err !=nil{
			//	continue
			//}
			//isSub := indexer.Capabilities().Categories.ContainsCat(categories.Subtitle)
			//if isSub{
			//	//This is a subtitle category
			//}
		}
	}

	agg := NewAggregate(indexers)
	return agg, nil
}

//CreateIndexer creates a new Indexer or aggregate Indexer with the given configuration.
func CreateIndexer(config config.Config, indexerName string) (Indexer, error) {
	def, err := DefaultDefinitionLoader.Load(indexerName)
	if err != nil {
		log.WithError(err).Warnf("Failed to load definition for %q. %v", indexerName, err)
		return nil, err
	}

	log.WithFields(log.Fields{"Indexer": indexerName}).Debugf("Loaded Indexer")
	indexer := NewRunner(def, RunnerOpts{
		Config: config,
	})
	return indexer, nil
}
