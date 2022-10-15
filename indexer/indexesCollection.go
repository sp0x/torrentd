package indexer

import (
	"errors"
	"strings"

	"github.com/sp0x/torrentd/indexer/search"
	"golang.org/x/sync/errgroup"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/categories"
)

//go:generate mockgen -source indexesCollection.go -destination=creation_mocks.go -package=indexer
type Scope interface {
	Lookup(config config.Config, key string) (IndexCollection, error)
	LookupWithCategories(config config.Config, selector *Selector, cats []categories.Category) (IndexCollection, error)
	LookupAll(config config.Config, selector *Selector) (IndexCollection, error)
	Indexes() map[string]IndexCollection
}

type IndexCollection []Indexer

func (i IndexCollection) Name() string {
	indexNames := make([]string, len(i))
	for i, ixr := range i {
		indexNames[i] = ixr.GetDefinition().Name
	}
	return strings.Join(indexNames, ",")
}

// HealthCheck checks all indexMap, if they can be searched.
func (i IndexCollection) HealthCheck() error {
	errorGroup := errgroup.Group{}
	for _, ixr := range i {
		indexerID := ixr.Info().GetID()
		// Run the Indexes in a goroutine
		errorGroup.Go(func() error {
			err := ixr.HealthCheck()
			if err != nil {
				log.Warnf("Indexes %q failed: %s", indexerID, err)
				return nil
			}
			return nil
		})
	}
	if err := errorGroup.Wait(); err != nil {
		log.Warn(err)
		return err
	}
	return nil
}

// Open try to open a scraping item from a collection of indexes
func (i IndexCollection) Open(scrapeItem search.ResultItemBase) (*ResponseProxy, error) {
	// Find the Indexes
	scrapeItemRoot := scrapeItem.AsScrapeItem()
	for _, index := range i {
		nfo := index.Info()
		if nfo.GetTitle() == scrapeItemRoot.Site {
			return index.Open(scrapeItem)
		}
	}
	return nil, errors.New("couldn't find Indexes")
}

func (i IndexCollection) Errors() []string {
	var allErrors []string
	for _, index := range i {
		allErrors = append(allErrors, index.Errors()...)
	}
	return allErrors
}

// MaxSearchPages returns the maximum number of pages that this aggregate can search, this is using the maximum paged index in the aggregate.
func (i IndexCollection) MaxSearchPages() uint {
	maxValue := uint(0)
	for _, index := range i {
		if index.MaxSearchPages() > maxValue {
			maxValue = index.MaxSearchPages()
		}
	}
	return maxValue
}

func (i IndexCollection) HasCategories(categories []categories.Category) bool {
	for _, index := range i {
		if index.Capabilities().HasCategories(categories) {
			return true
		}
	}
	return false
}

type indexMap struct {
	indexes map[string]IndexCollection
	loader  DefinitionLoader
}

// NewScope creates a new scope for indexes that can be or are loaded
func NewScope(definitionLoader DefinitionLoader) Scope {
	sc := &indexMap{}
	sc.indexes = make(map[string]IndexCollection)
	if definitionLoader == nil {
		definitionLoader = GetIndexDefinitionLoader()
	}
	sc.loader = definitionLoader
	return sc
}

// Indexes returns the currently loaded indexMap
func (c *indexMap) Indexes() map[string]IndexCollection {
	return c.indexes
}

// Lookup finds the matching Indexer.
func (c *indexMap) Lookup(config config.Config, indexSelectionKey string) (IndexCollection, error) {
	// If we already have that indexer running, we don't create a new one.
	selector := newIndexSelector(indexSelectionKey)
	log.Debugf("Looking up scoped index: %v\n", selector)
	if _, ok := c.indexes[indexSelectionKey]; !ok {
		var indexes []Indexer
		var err error
		// If we're looking up an aggregate indexes, we just create an aggregate
		if selector.isAggregate() {
			indexes, err = c.LookupAll(config, selector)
		} else {
			indexes, err = NewIndexRunnerByNameOrSelector(selector.Value(), config)
		}
		if err != nil {
			return nil, err
		}
		c.indexes[indexSelectionKey] = indexes
	}
	return c.indexes[indexSelectionKey], nil
}

// LookupWithCategories creates a new aggregate with the indexMap that match a set of indexCategories
func (c *indexMap) LookupWithCategories(config config.Config, selector *Selector, cats []categories.Category) (IndexCollection, error) {
	indexKeys, err := c.loader.ListAvailableIndexes(selector)
	if err != nil {
		return nil, err
	}

	var indexes IndexCollection
	for _, key := range indexKeys {
		currentIndexes, err := c.Lookup(config, key)
		if err != nil {
			return nil, err
		}
		for _, ix := range currentIndexes {
			if !ix.Capabilities().HasCategories(cats) {
				continue
			}
			indexes = append(indexes, ix)
		}
	}
	return indexes, nil
}

// LookupAll creates an aggregate of all the valid configured indexMap
// this includes indexMap that don't need a login.
func (c *indexMap) LookupAll(config config.Config, selector *Selector) (IndexCollection, error) {
	var keysToLoad []string
	var err error
	keysToLoad, err = c.loader.ListAvailableIndexes(selector)
	if err != nil {
		return nil, err
	}
	if keysToLoad == nil {
		log.WithFields(log.Fields{"selector": selector, "loader": c.loader}).
			Debug("Tried to create an aggregate index where no child indexMap could be found")
		return nil, errors.New("no indexMap matched the given selector")
	}

	//result := &Aggregate{}
	var indexes []Indexer
	if selector != nil {
		//selectorCopy := *selector
		//result.selector = &selectorCopy
	}
	for _, key := range keysToLoad {
		// Search the site configuration, we only use configured indexMap
		indexConfig, _ := config.GetSite(key) // Search all the configured indexMap
		if indexConfig != nil {
			index, err := c.Lookup(config, key)
			if err != nil {
				return nil, err
			}
			//result.Indexes = append(result.Indexes, index)
			indexes = append(indexes, index...)
		} else {
			log.WithFields(log.Fields{"index": key}).
				Debug("Tried to load an index that has no config")
		}
	}
	if len(indexes) == 0 {
		log.WithFields(log.Fields{"selector": selector}).
			Debug("Created aggregate without any indexMap")
	}
	return indexes, nil
}

type IndexCollectionInfo struct{}

func (a *IndexCollectionInfo) GetLanguage() string {
	return "en-US"
}

func (a *IndexCollectionInfo) GetLink() string {
	return ""
}

func (a *IndexCollectionInfo) GetTitle() string {
	return "Index collection"
}

func (a *IndexCollectionInfo) GetID() string {
	return aggregateSiteName
}

func (i IndexCollection) Info() Info {
	return &IndexCollectionInfo{}
}
