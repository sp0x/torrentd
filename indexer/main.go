package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
)

var indexers map[string]Indexer

func init() {
	indexers = make(map[string]Indexer)
}

func Lookup(config config.Config, key string) (Indexer, error) {
	if key == "aggregate" || key == "all" {
		return CreateAggregate(config)
	}
	if _, ok := indexers[key]; !ok {
		indexer, err := CreateIndexer(config, key)
		if err != nil {
			return nil, err
		}
		indexers[key] = indexer
	}
	return indexers[key], nil
}

func CreateAggregate(config config.Config) (Indexer, error) {
	keys, err := DefaultDefinitionLoader.List()
	if err != nil {
		return nil, err
	}

	agg := Aggregate{}
	for _, key := range keys {
		ifaceConfig, _ := config.GetSite(key) //Get all the configured indexers
		if ifaceConfig != nil && len(ifaceConfig) > 0 {
			indexer, err := Lookup(config, key)
			if err != nil {
				return nil, err
			}
			agg = append(agg, indexer)
		}
	}

	return agg, nil
}

//CreateIndexer creates a new indexer or aggregate indexer with the given configuration.
func CreateIndexer(config config.Config, indexerName string) (Indexer, error) {
	def, err := DefaultDefinitionLoader.Load(indexerName)
	if err != nil {
		log.WithError(err).Warnf("Failed to load definition for %q. %v", indexerName, err)
		return nil, err
	}

	log.WithFields(log.Fields{"indexer": indexerName}).Debugf("Loaded indexer")
	indexer := NewRunner(def, RunnerOpts{
		Config: config,
	})
	return indexer, nil
}
