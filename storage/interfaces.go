package storage

import (
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
)

type ItemStorage interface {
	Size() int64
	Find(query indexing.Query, output *search.ExternalResultItem) error
	Add(item search.Record) error
	AddUniqueIndex(key *indexing.Key)
	NewWithKey(pk *indexing.Key) ItemStorage
	Close()
	SetKey(index *indexing.Key) error
	GetNewest(count int) []search.ExternalResultItem
	ForEach(callback func(record interface{}))
}
type ItemStorageBacking interface {
	//Tries to find a single record matching the query.
	Find(query indexing.Query, result interface{}) error
	HasIndex(meta *indexing.IndexMetadata) bool
	GetIndexes() map[string]indexing.IndexMetadata
	Update(query indexing.Query, item interface{}) error
	//CreateWithId creates a new record using a custom key
	CreateWithId(parts *indexing.Key, item search.Record, uniqueIndexKeys *indexing.Key) error
	//Create a new record with the default key (UUIDValue)
	Create(item search.Record, additionalPK *indexing.Key) error
	//Size is the size of the storage, as in records count
	Size() int64
	//GetNewest returns the latest `count` of records.
	GetNewest(count int) []search.ExternalResultItem
	Close()
	ForEach(callback func(record interface{}))
}
