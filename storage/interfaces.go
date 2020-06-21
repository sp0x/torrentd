package storage

import (
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
)

type ItemStorage interface {
	Add(item *search.ExternalResultItem) (bool, bool)
	NewWithKey(key indexing.Key) ItemStorage
}
type ItemStorageBacking interface {
	//Tries to find a single record matching the query.
	Find(query indexing.Query, result *search.ExternalResultItem) error
	Update(query indexing.Query, item *search.ExternalResultItem) error
	Create(parts indexing.Key, item *search.ExternalResultItem) error
}
