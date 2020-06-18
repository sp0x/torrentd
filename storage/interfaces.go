package storage

import "github.com/sp0x/torrentd/indexer/search"

type ItemStorage interface {
	Add(item *search.ExternalResultItem) (bool, bool)
}
type ItemStorageBacking interface {
	Find(key Query) *search.ExternalResultItem
	Update(key Query, item *search.ExternalResultItem)
	Create(item *search.ExternalResultItem)
}
