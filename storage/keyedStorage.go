package storage

import (
	"errors"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/storage/stats"
)

type KeyedStorage struct {
	backing        ItemStorageBacking
	primaryKey     indexing.Key
	indexKeys      indexing.Key
	indexKeysCache map[string]interface{}
}

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

func (s *KeyedStorage) GetLatest(count int) []search.ResultItemBase {
	return s.backing.GetLatest(count)
}

func (s *KeyedStorage) Size() int64 {
	return s.backing.Size()
}

func (s *KeyedStorage) getDefaultKey() *indexing.Key {
	//Use the ID from the result as a key
	key := indexing.NewKey("UUID")
	return key
}

func (s *KeyedStorage) Find(query indexing.Query, output *search.ScrapeResultItem) error {
	if s.backing.Find(query, output) == nil {
		return nil
	}
	return errors.New("not found")
}

func (s *KeyedStorage) ForEach(callback func(record search.ResultItemBase)) {
	s.backing.ForEach(callback)
}

func (s *KeyedStorage) GetStats(showDebugInfo bool) *stats.Stats {
	return s.backing.GetStats(showDebugInfo)
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
	var existingResult *search.ScrapeResultItem
	var existingQuery indexing.Query
	//The key is what makes each result unique. If no key is provided you might end up with doubles, since UUIDValue is used.
	key := &s.primaryKey
	if key.IsEmpty() {
		key = s.getDefaultKey()
	}
	keyHasValue := indexing.KeyHasValue(key, item)
	if keyHasValue {
		existingQuery = indexing.GetKeyQueryFromItem(key, item)
		if existingQuery != nil {
			tmpResult := search.ScrapeResultItem{}
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
