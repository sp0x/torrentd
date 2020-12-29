package indexer

import (
	"github.com/sp0x/torrentd/storage/indexing"
)

// entityBlock describes an entity data type that's present in an index.
type entityBlock struct {
	Name     string        `yaml:"name"`
	IndexKey stringorslice `yaml:"key"`
}

// GetKey gets the indexing key for this entity.
func (b entityBlock) GetKey() *indexing.Key {
	key := indexing.NewKey()
	for _, s := range b.IndexKey {
		key.Add(s)
	}
	return key
}
