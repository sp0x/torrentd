package storage

import "github.com/sp0x/torrentd/indexer/search"

type ItemStorage interface {
	Add(item *search.ExternalResultItem) (bool, bool)
	NewWithKey(key Key) ItemStorage
}
type ItemStorageBacking interface {
	//Tries to find a single record matching the query.
	Find(query Query, result *search.ExternalResultItem) error
	Update(query Query, item *search.ExternalResultItem) error
	Create(parts Key, item *search.ExternalResultItem) error
}
