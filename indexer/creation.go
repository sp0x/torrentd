package indexer

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
)

//go:generate mockgen -source creation.go -destination=creation_mocks_test.go -package=indexer
type Scope interface {
	Lookup(config config.Config, key string) (Indexer, error)
	CreateAggregateForCategories(config config.Config, selector *Selector, cats []categories.Category) (Indexer, error)
	CreateAggregate(config config.Config, selector *Selector) (Indexer, error)
	Indexes() map[string]Indexer
}

type CachedScope struct {
	indexes map[string]Indexer
}

// NewScope creates a new scope for indexer runners
func NewScope() Scope {
	sc := &CachedScope{}
	sc.indexes = make(map[string]Indexer)
	return sc
}

// Indexes returns the currently loaded indexes
func (c *CachedScope) Indexes() map[string]Indexer {
	return c.indexes
}

// Lookup finds the matching Indexer.
func (c *CachedScope) Lookup(config config.Config, key string) (Indexer, error) {
	// If we already have that indexer running, we don't create a new one.
	selector := newIndexerSelector(key)
	log.Debugf("Looking up scoped index: %v\n", selector)
	if _, ok := c.indexes[key]; !ok {
		var indexer Indexer
		var err error
		// If we're looking up an aggregate indexer, we just create an aggregate
		if selector.isAggregate() {
			indexer, err = c.CreateAggregate(config, selector)
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

// CreateAggregateForCategories creates a new aggregate with the indexes that match a set of indexCategories
func (c *CachedScope) CreateAggregateForCategories(config config.Config, selector *Selector, cats []categories.Category) (Indexer, error) {
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

// CreateAggregate creates an aggregate of all the valid configured indexes
// this includes indexes that don't need a login.
func (c *CachedScope) CreateAggregate(config config.Config, selector *Selector) (Indexer, error) {
	var keysToLoad []string
	var err error
	keysToLoad, err = Loader.List(selector)
	if err != nil {
		return nil, err
	}
	if keysToLoad == nil {
		log.WithFields(log.Fields{"selector": selector, "loader": Loader}).
			Debug("Tried to create an aggregate index where no child indexes could be found")
		return nil, errors.New("no indexes matched the given selector")
	}

	result := &Aggregate{}
	if selector != nil {
		selectorCopy := *selector
		result.selector = &selectorCopy
	}
	for _, key := range keysToLoad {
		// Get the site configuration, we only use configured indexes
		indexConfig, _ := config.GetSite(key) // Get all the configured indexes
		if indexConfig != nil {
			index, err := c.Lookup(config, key)
			if err != nil {
				return nil, err
			}
			result.Indexers = append(result.Indexers, index)
		} else {
			log.WithFields(log.Fields{"index": key}).
				Debug("Tried to load an index that has no config")
		}
	}
	if result.Indexers == nil {
		log.WithFields(log.Fields{"selector": selector}).
			Debug("Created aggregate without any indexes")
	}
	return result, nil
}

// CreateIndexer creates a new Indexer or aggregate Indexer with the given configuration.
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
