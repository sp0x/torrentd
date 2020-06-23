package storage

import (
	"github.com/prometheus/common/log"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/firebase"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/spf13/viper"
)

type KeyedStorage struct {
	//backing *DBStorage
	backing        ItemStorageBacking
	primaryKey     indexing.Key
	indexKeys      indexing.Key
	indexKeysCache map[string]interface{}
}

//NewKeyedStorage creates a new keyed storage with the default storage backing.
func NewKeyedStorage(keyFields *indexing.Key) *KeyedStorage {
	return &KeyedStorage{
		primaryKey:     *keyFields,
		backing:        DefaultStorageBacking(),
		indexKeysCache: make(map[string]interface{}),
	}
}

//DefaultStorageBacking gets the default storage method for results.
func DefaultStorageBacking() ItemStorageBacking {
	backing, err := bolt.NewBoltStorage("")
	if err != nil {
		panic(err)
	}
	return backing
}

//NewKeyedStorageWithBacking creates a new keyed storage with a custom storage backing.
func NewKeyedStorageWithBacking(key *indexing.Key, storage ItemStorageBacking) *KeyedStorage {
	return &KeyedStorage{
		primaryKey:     *key,
		backing:        storage,
		indexKeysCache: make(map[string]interface{}),
	}
}

func NewKeyedStorageWithBackingType(namespace string, key *indexing.Key, storageType string) *KeyedStorage {
	bfn, ok := storageBackingMap[storageType]
	if !ok {
		panic("Unsupported storage backing type")
	}
	b := bfn(namespace)
	return NewKeyedStorageWithBacking(key, b)
}

var storageBackingMap = make(map[string]func(ns string) ItemStorageBacking)

func init() {
	storageBackingMap["boltdb"] = func(ns string) ItemStorageBacking {
		b, err := bolt.NewBoltStorage("")
		b.SetNamespace(ns)
		if err != nil {
			log.Error(err)
			return nil
		}
		return b
	}
	storageBackingMap["firebase"] = func(ns string) ItemStorageBacking {
		conf := &firebase.FirestoreConfig{Namespace: ns}
		conf.ProjectId = viper.Get("firebase_project").(string)
		conf.CredentialsFile = viper.Get("firebase_credentials_file").(string)
		b, err := firebase.NewFirestoreStorage(conf)
		if err != nil {
			log.Error(err)
			return nil
		}
		return b
	}
	storageBackingMap["sqlite"] = func(ns string) ItemStorageBacking {
		panic("Deprecated")
		//b := &sqlite.DBStorage{}
		//return b
	}
}

//NewWithKey gets a storage backed in the same way, with a different key.
func (s *KeyedStorage) NewWithKey(key *indexing.Key) ItemStorage {
	storage := s.backing

	return &KeyedStorage{
		primaryKey: *key,
		backing:    storage,
	}
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

//Add handles the discovery of the result, adding additional information like staleness state.
func (s *KeyedStorage) Add(item *search.ExternalResultItem) error {
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
		item.Fingerprint = search.GetResultFingerprint(item)
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
		item.Fingerprint = existingResult.Fingerprint
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
