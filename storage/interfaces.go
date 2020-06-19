package storage

import "github.com/sp0x/torrentd/indexer/search"

type ItemStorage interface {
	Add(item *search.ExternalResultItem) (bool, bool)
}
type ItemStorageBacking interface {
	Find(query Query, result *search.ExternalResultItem) error
	Update(query Query, item *search.ExternalResultItem) error
	Create(parts Key, item *search.ExternalResultItem) error
}
