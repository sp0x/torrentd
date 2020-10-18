package indexer

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
	"strings"
)

//go:generate mockgen -source creation.go -destination=mocks/creation.go -package=mocks
type Scope interface {
	Lookup(config config.Config, key string) (Indexer, error)
	CreateAggregateForCategories(config config.Config, selector *IndexerSelector, cats []categories.Category) (Indexer, error)
	CreateAggregate(config config.Config, selector *IndexerSelector) (Indexer, error)
}

type CachedScope struct {
	indexers map[string]Indexer
}

//NewScope creates a new scope for indexer runners
func NewScope() Scope {
	sc := &CachedScope{}
	sc.indexers = make(map[string]Indexer)
	return sc
}

func (s IndexerSelector) isAggregate() bool {
	return s.selector == "" || s.selector == "aggregate" || s.selector == "all" || strings.Contains(s.selector, ",")
}

type IndexerSelector struct {
	selector string
	parts    []string
}

func (s IndexerSelector) shouldLoadAllIndexes() bool {
	indexKeys := strings.Split(s.selector, ",")
	return s.isAggregate() && len(indexKeys) == 1
}

func (s IndexerSelector) Matches(name string) bool {
	if s.selector == "" || s.selector == "all" || s.selector == "aggregate" {
		return false
	}
	return contains(s.parts, name)
}

func (s IndexerSelector) Value() string {
	return s.selector
}

func newIndexerSelector(selector string) IndexerSelector {
	return IndexerSelector{selector: selector, parts: strings.Split(selector, ",")}
}

//Lookup finds the matching Indexer.
func (c *CachedScope) Lookup(config config.Config, key string) (Indexer, error) {
	//If we already have that indexer running, we don't create a new one.
	selector := newIndexerSelector(key)
	if _, ok := c.indexers[key]; !ok {
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
		c.indexers[key] = indexer
	}
	return c.indexers[key], nil
}

//CreateAggregateForCategories creates a new aggregate with the indexers that match a set of categories
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

//CreateAggregate creates an aggregate of all the valid configured indexers
//this includes indexers that don't need a login.
func (c *CachedScope) CreateAggregate(config config.Config, selector *IndexerSelector) (Indexer, error) {
	var keysToLoad []string = nil
	var err error
	keysToLoad, err = Loader.List(selector)
	if err != nil {
		return nil, err
	}

	result := &Aggregate{}
	for _, key := range keysToLoad {
		//Get the site configuration, we only use configured indexers
		ifaceConfig, _ := config.GetSite(key) //Get all the configured indexers
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
