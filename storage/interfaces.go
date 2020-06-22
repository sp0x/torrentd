package storage

import (
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
)

type ItemStorage interface {
	Add(item *search.ExternalResultItem) error
	NewWithKey(pk indexing.Key) ItemStorage
}
type ItemStorageBacking interface {
	//Tries to find a single record matching the query.
	Find(query indexing.Query, result *search.ExternalResultItem) error
	Update(query indexing.Query, item *search.ExternalResultItem) error
	//CreateWithId creates a new record using a custom key
	CreateWithId(parts indexing.Key, item *search.ExternalResultItem, uniqueIndexKeys indexing.Key) error
	//Create a new record with the default key (GUID)
	Create(item *search.ExternalResultItem, additionalPK indexing.Key) error
	//Size is the size of the storage, as in records count
	Size() int64
	//GetNewest returns the latest `count` of records.
	GetNewest(count int) []search.ExternalResultItem
}
