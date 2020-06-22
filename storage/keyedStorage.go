package storage

import (
	"github.com/prometheus/common/log"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/firebase"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/storage/sqlite"
	"github.com/spf13/viper"
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

//DefaultStorageBacking gets the default storage method for results.
func DefaultStorageBacking() ItemStorageBacking {
	backing, err := bolt.NewBoltStorage("")
	if err != nil {
		panic(err)
	}
	return backing
}

//NewKeyedStorageWithBacking creates a new keyed storage with a custom storage backing.
func NewKeyedStorageWithBacking(key indexing.Key, storage ItemStorageBacking) *KeyedStorage {
	return &KeyedStorage{
		keyParts: key,
		backing:  storage,
	}
}

func NewKeyedStorageWithBackingType(key indexing.Key, storageType string) *KeyedStorage {
	bfn, ok := storageBackingMap[storageType]
	if !ok {
		panic("Unsupported storage backing type")
	}
	b := bfn()
	return NewKeyedStorageWithBacking(key, b)
}

var storageBackingMap = make(map[string]func() ItemStorageBacking)

func init() {
	storageBackingMap["boltdb"] = func() ItemStorageBacking {
		b, err := bolt.NewBoltStorage("")
		if err != nil {
			log.Error(err)
			return nil
		}
		return b
	}
	storageBackingMap["firebase"] = func() ItemStorageBacking {
		conf := &firebase.FirestoreConfig{}
		conf.ProjectId = viper.Get("firebase_project").(string)
		conf.CredentialsFile = viper.Get("firebase_credentials_file").(string)
		b, err := firebase.NewFirestoreStorage(conf)
		if err != nil {
			log.Error(err)
			return nil
		}
		return b
	}
	storageBackingMap["sqlite"] = func() ItemStorageBacking {
		b := &sqlite.DBStorage{}
		return b
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

func (s *KeyedStorage) GetNewest(count int) []search.ExternalResultItem {
	return s.backing.GetNewest(count)
}

func (s *KeyedStorage) Size() int64 {
	return s.backing.Size()
}

func (s *KeyedStorage) getDefaultKey() indexing.Key {
	//Use the ID from the result as a key
	key := indexing.Key{}
	key = append(key, "ID")
	return key
}

//Add handles the discovery of the result, adding additional information like staleness state.
func (s *KeyedStorage) Add(item *search.ExternalResultItem) (bool, bool) {
	var existingResult *search.ExternalResultItem
	key := s.keyParts
	//We let the backing deal with the ID/GUID
	//if key == nil {
	//	key = s.getDefaultKey()
	//}
	existingKey := indexing.GetKeyQueryFromItem(key, item)
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
		err := s.backing.Create(key, item)
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
