package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
)

//go:generate mockgen -source creation.go -destination=creation_mocks_test.go -package=indexer
type Scope interface {
	Lookup(config config.Config, key string) (Indexer, error)
	CreateAggregateForCategories(config config.Config, selector *IndexerSelector, cats []categories.Category) (Indexer, error)
	CreateAggregate(config config.Config, selector *IndexerSelector) (Indexer, error)
	Indexes() map[string]Indexer
}

type CachedScope struct {
	indexes map[string]Indexer
}

//NewScope creates a new scope for indexer runners
func NewScope() Scope {
	sc := &CachedScope{}
	sc.indexes = make(map[string]Indexer)
	return sc
}

//Indexes returns the currently loaded indexes
func (c *CachedScope) Indexes() map[string]Indexer {
	return c.indexes
}

//Lookup finds the matching Indexer.
func (c *CachedScope) Lookup(config config.Config, key string) (Indexer, error) {
	//If we already have that indexer running, we don't create a new one.
	selector := newIndexerSelector(key)
	if _, ok := c.indexes[key]; !ok {
		var indexer Indexer
		var err error
		//If we're looking up an aggregate indexer, we just create an aggregate
		if selector.isAggregate() {
			indexer, err = c.CreateAggregate(config, &selector)
		} else {
			indexer, err = CreateIndexer(config, selector.Value())
		}
		if err != nil {
			return nil, err
		}
		c.indexes[key] = indexer
	}
	return c.indexes[key], nil
}

//CreateAggregateForCategories creates a new aggregate with the indexes that match a set of categories
func (c *CachedScope) CreateAggregateForCategories(config config.Config, selector *IndexerSelector, cats []categories.Category) (Indexer, error) {
	ixrKeys, err := Loader.List(selector)
	if err != nil {
		return nil, err
	}

	result := &Aggregate{}
	for _, key := range ixrKeys {
		ixr, err := c.Lookup(config, key)
		if err != nil {
			return nil, err
		}
		if !ixr.Capabilities().HasCategories(cats) {
			continue
		}
		result.Indexers = append(result.Indexers, ixr)
	}
	return result, nil
}

//CreateAggregate creates an aggregate of all the valid configured indexes
//this includes indexes that don't need a login.
func (c *CachedScope) CreateAggregate(config config.Config, selector *IndexerSelector) (Indexer, error) {
	var keysToLoad []string = nil
	var err error
	keysToLoad, err = Loader.List(selector)
	if err != nil {
		return nil, err
	}

	result := &Aggregate{}
	for _, key := range keysToLoad {
		//Get the site configuration, we only use configured indexes
		ifaceConfig, _ := config.GetSite(key) //Get all the configured indexes
		if ifaceConfig != nil {
			indexer, err := c.Lookup(config, key)
			if err != nil {
				return nil, err
			}
			result.Indexers = append(result.Indexers, indexer)
		}
		//else {
		//Indexer might not be configured
		//indexer, err := Lookup(config, key)
		//if err !=nil{
		//	continue
		//}
		//isSub := indexer.Capabilities().Categories.ContainsCat(categories.Subtitle)
		//if isSub{
		//	//This is a subtitle category
		//}
		//}
	}
	return result, nil
}

//CreateIndexer creates a new Indexer or aggregate Indexer with the given configuration.
func CreateIndexer(config config.Config, indexerName string) (Indexer, error) {
	def, err := Loader.Load(indexerName)
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
