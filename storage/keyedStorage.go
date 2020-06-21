package storage

import (
	"github.com/prometheus/common/log"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
)

type KeyedStorage struct {
	//backing *DBStorage
	backing  ItemStorageBacking
	keyParts indexing.Key
}

//NewKeyedStorage creates a new keyed storage with the default storage backing.
func NewKeyedStorage(keyFields indexing.Key) *KeyedStorage {
	return &KeyedStorage{
		keyParts: keyFields,
		backing:  DefaultStorageBacking(),
	}
}

//NewKeyedStorageWithBacking creates a new keyed storage with a custom storage backing.
func NewKeyedStorageWithBacking(key indexing.Key, storage ItemStorageBacking) *KeyedStorage {
	return &KeyedStorage{
		keyParts: key,
		backing:  storage,
	}
}

//NewWithKey gets a storage backed in the same way, with a different key.
func (s *KeyedStorage) NewWithKey(key indexing.Key) ItemStorage {
	storage := s.backing

	return &KeyedStorage{
		keyParts: key,
		backing:  storage,
	}
}

//Add handles the discovery of the result, adding additional information like staleness state.
func (s *KeyedStorage) Add(item *search.ExternalResultItem) (bool, bool) {
	var existingResult *search.ExternalResultItem
	existingKey := indexing.GetKeyQueryFromItem(s.keyParts, item)
	if existingKey != nil {
		tmpResult := search.ExternalResultItem{}
		if s.backing.Find(existingKey, &tmpResult) == nil {
			existingResult = &tmpResult
		}
	}
	isNew := false
	isUpdate := false
	if existingResult == nil {
		isNew = true
		item.Fingerprint = search.GetResultFingerprint(item)
		err := s.backing.Create(s.keyParts, item)
		if err != nil {
			log.Error(err)
			return false, false
		}
	} else if !existingResult.Equals(item) {
		//This must be an update
		isUpdate = true
		item.Fingerprint = existingResult.Fingerprint
		err := s.backing.Update(existingKey, item)
		if err != nil {
			log.Error(err)
			return false, false
		}
	}
	//We set the result's state so it's known later on whenever it's used.
	item.SetState(isNew, isUpdate)
	return isNew, isUpdate
}
