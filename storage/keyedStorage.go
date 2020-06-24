package storage

import (
	"errors"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/indexing"
)

type KeyedStorage struct {
	//backing *DBStorage
	backing        ItemStorageBacking
	primaryKey     indexing.Key
	indexKeys      indexing.Key
	indexKeysCache map[string]interface{}
}

//NewKeyedStorage creates a new keyed storage with the default storage backing.
//func NewKeyedStorage(keyFields *indexing.Key) *KeyedStorage {
//	return &KeyedStorage{
//		primaryKey:     *keyFields,
//		backing:        DefaultStorageBacking(),
//		indexKeysCache: make(map[string]interface{}),
//	}
//}

//DefaultStorageBacking gets the default storage method for results.
func DefaultStorageBacking() ItemStorageBacking {
	backing, err := bolt.NewBoltStorage("")
	if err != nil {
		panic(err)
	}
	return backing
}

//NewKeyedStorageWithBacking creates a new keyed storage with a custom storage backing.
//func NewKeyedStorageWithBacking(key *indexing.Key, storage ItemStorageBacking) *KeyedStorage {
//
//}

//func NewKeyedStorageWithBackingType(namespace string, config config.Config, key *indexing.Key, storageType string) *KeyedStorage {
//
//}

//NewWithKey gets a storage backed in the same way, with a different key.
func (s *KeyedStorage) NewWithKey(key *indexing.Key) ItemStorage {
	storage := s.backing
	return &KeyedStorage{
		primaryKey: *key,
		backing:    storage,
	}
}

func (s *KeyedStorage) Close() {
	s.backing.Close()
}

func (s *KeyedStorage) GetNewest(count int) []search.ExternalResultItem {
	return s.backing.GetNewest(count)
}

func (s *KeyedStorage) Size() int64 {
	return s.backing.Size()
}

func (s *KeyedStorage) getDefaultKey() *indexing.Key {
	//Use the ID from the result as a key
	key := indexing.NewKey("GUID")
	return key
}

func (s *KeyedStorage) Find(query indexing.Query, output *search.ExternalResultItem) error {
	if s.backing.Find(query, output) == nil {
		return nil
	}
	return errors.New("not found")
}

func (s *KeyedStorage) ForEach(callback func(record interface{})) {
	panic("implement me")
}

func (s *KeyedStorage) SetKey(index *indexing.Key) error {
	if index.IsEmpty() {
		return errors.New("primary key was empty")
	}
	s.primaryKey = *index
	return nil
}

//Add handles the discovery of the result, adding additional information like staleness state.
func (s *KeyedStorage) Add(item search.Record) error {
	var existingResult *search.ExternalResultItem
	var existingQuery indexing.Query
	//The key is what makes each result unique. If no key is provided you might end up with doubles, since GUID is used.
	key := &s.primaryKey
	if key.IsEmpty() {
		key = s.getDefaultKey()
	}
	keyHasValue := indexing.KeyHasValue(key, item)
	if keyHasValue {
		existingQuery = indexing.GetKeyQueryFromItem(key, item)
		if existingQuery != nil {
			tmpResult := search.ExternalResultItem{}
			if s.backing.Find(existingQuery, &tmpResult) == nil {
				existingResult = &tmpResult
			}
		}
	}
	isNew := false
	isUpdate := false
	if existingResult == nil {
		isNew = true
		var err error
		if keyHasValue {
			err = s.backing.CreateWithId(key, item, &s.indexKeys)
		} else {
			err = s.backing.Create(item, &s.indexKeys)
		}
		if err != nil {
			return err
		}
	} else if !existingResult.Equals(item) {
		//This must be an update
		isUpdate = true
		err := s.backing.Update(existingQuery, item)
		if err != nil {
			return err
		}
	}
	//We set the result's state so it's known later on whenever it's used.
	item.SetState(isNew, isUpdate)
	return nil
}

//AddUniqueIndex adds a new key set as unique for this storage.
func (s *KeyedStorage) AddUniqueIndex(key *indexing.Key) {
	s.indexKeys.AddKeys(key)
}
