package bolt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/storage/serializers"
	"github.com/sp0x/torrentd/storage/serializers/json"
	"os"
	"path"
	"time"
)

const (
	resultsBucket = "results"
)

var categoriesInitialized = false

type BoltStorage struct {
	Database   *bolt.DB
	rootBucket []string
	marshaler  *serializers.DynamicMarshaler
	metadata   *Metadata
}

func ensurePathExists(dbPath string) {
	if dbPath == "" {
		return
	}
	//if !strings.HasSuffix(dbPath, ".db") && !strings.HasSuffix(dbPath, "/") {
	//	dbPath += "/"
	//}
	dirPath := path.Dir(dbPath)
	_ = os.MkdirAll(dirPath, os.ModePerm)
}

func NewBoltStorage(dbPath string, recordTypePtr interface{}) (*BoltStorage, error) {
	if dbPath == "" {
		dbPath = DefaultBoltPath()
	}
	ensurePathExists(dbPath)
	dbx, err := GetBoltDb(dbPath)
	if err != nil {
		return nil, err
	}
	bls := &BoltStorage{
		Database:  dbx,
		marshaler: serializers.NewDynamicMarshaler(recordTypePtr, json.Serializer),
	}
	err = bls.setupMetadata()
	if err != nil {
		bls.Close()
		return nil, err
	}
	return bls, nil
}

func GetBoltDb(file string) (*bolt.DB, error) {
	dbx, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	//Setup our DB
	err = dbx.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("searchResults"))
		if err != nil {
			return err
		}

		//_, err = tx.CreateBucketIfNotExists([]byte("telegram_chats"))
		//if err != nil {
		//	return err
		//}
		//CreateWithId all of our categories
		if !categoriesInitialized {
			for _, cat := range categories.AllCategories {
				catKey := []byte(cat.Name)
				_, err := root.CreateBucketIfNotExists(catKey)
				if err != nil {
					return err
				}
			}
		}
		categoriesInitialized = true
		return err
	})
	if err != nil {
		return nil, err
	}
	return dbx, nil
}

func (b *BoltStorage) Close() {
	if b.Database == nil {
		return
	}
	_ = b.Database.Close()
}

//Find something by it's index keys.
//Todo: refactor this
func (b *BoltStorage) Find(query indexing.Query, result interface{}) error {
	if query == nil {
		return errors.New("query is required")
	}
	indexValue := indexing.GetIndexValueFromQuery(query)
	extract := func(bucket *bolt.Bucket, idx indexing.Index) error {
		ids := idx.All(indexValue, indexing.SingleItemCursor())
		if len(ids) == 0 {
			return errors.New("not found")
		}
		//Use the same root bucket and query by the ID in our index
		rawResult := bucket.Get(ids[0])
		var err error
		//If the result is nil, we night be using the additional PK
		if rawResult == nil {
			//TODO: add the pk information into the root bucket, so we know if we need to do this.
			var secondaryIndex indexing.Index
			secondaryIndex, err = b.GetUniqueIndexFromKeys(bucket, indexing.NewKey("UUID"))
			if err != nil {
				return err
			}
			ids := secondaryIndex.Get(ids[0])
			rawResult = bucket.Get(ids)
		}
		err = b.marshaler.UnmarshalAt(rawResult, result)
		if err != nil {
			return err
		}
		return nil
	}
	//The our bucket, and the index that matches the query best
	err := b.Database.View(func(tx *bolt.Tx) error {
		bucket := b.GetBucket(tx, resultsBucket)
		if bucket == nil {
			return errors.New("not found")
		}
		idx, err := b.GetIndexFromQuery(bucket, query)
		if err != nil {
			return err
		}
		return extract(bucket, idx)
	})
	//At this point we can quit.
	if err == nil {
		return nil
	}
	//We should retry
	if _, ok := err.(*IndexDoesNotExistAndNotWritable); ok {
		err = b.Database.Update(func(tx *bolt.Tx) error {
			_, err := b.GetIndexFromQuery(b.GetBucket(tx, resultsBucket), query)
			return err
		})
		if err != nil {
			return err
		}
		err = b.Database.View(func(tx *bolt.Tx) error {
			bucket := b.GetBucket(tx, resultsBucket)
			if bucket == nil {
				return errors.New("not found")
			}
			idx, err := b.GetIndexFromQuery(bucket, query)
			if err != nil {
				return err
			}
			return extract(bucket, idx)
		})
		return err
	}
	return err
}

func (b *BoltStorage) Update(query indexing.Query, item interface{}) error {
	if query == nil {
		return errors.New("query is required")
	}
	return b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		idx, err := b.GetIndexFromQuery(bucket, query)
		if err != nil {
			return err
		}
		indexValue := indexing.GetIndexValueFromQuery(query)
		//Fetch the ID from the index
		ids := idx.All(indexValue, indexing.SingleItemCursor())
		//Serialize the item
		serializedValue, err := b.marshaler.Marshal(item)
		if err != nil {
			return err
		}
		//Put the serialized value in that place
		return bucket.Put(ids[0], serializedValue)
	})

}

//Create a new record. This uses a new random UUID in order to identify the record.
func (b *BoltStorage) Create(item search.Record, additionalPK *indexing.Key) error {
	item.SetUUID(uuid.New().String())
	key := indexing.NewKey("UUID")
	err := b.CreateWithId(key, item, nil)
	if err != nil {
		return err
	}
	//If we don't have an unique index, we can stop here.
	if additionalPK == nil || additionalPK.IsEmpty() {
		return nil
	}
	indexValue := indexing.GetIndexValueFromItem(additionalPK, item)
	//We need add a new index: additionalPK -> UUIDValue
	return b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		//We get the keyIndex that we'll use
		keyToGuidIndex, err := b.GetUniqueIndexFromKeys(bucket, additionalPK)
		if err != nil {
			return err
		}
		guidBytes := []byte(item.UUID())
		//Save the keyIndex for the id of the result.
		err = keyToGuidIndex.Add(indexValue, guidBytes)
		return err
	})
}

//CreateWithId a new record for a result.
//The key is used if you have a custom object that uses a different key, not the UUIDValue
func (b *BoltStorage) CreateWithId(keyParts *indexing.Key, item search.Record, uniqueIndexKeys *indexing.Key) error {
	indexValue := indexing.GetIndexValueFromItem(keyParts, item)
	uniqueIndexValue := indexing.GetIndexValueFromItem(uniqueIndexKeys, item)
	if len(uniqueIndexValue) == 0 {
		uniqueIndexValue = []byte("\000;\000")
	}
	return b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		//We get the pkIndex that we'll use
		pkIndex, err := b.GetUniqueIndexFromKeys(bucket, keyParts)
		if err != nil {
			return err
		}
		var uniqueIndex indexing.Index
		if uniqueIndexKeys != nil && !uniqueIndexKeys.IsEmpty() {
			uniqueIndex, err = b.GetUniqueIndexFromKeys(bucket, uniqueIndexKeys)
			if err != nil {
				return err
			}
			existingUniqueVal := uniqueIndex.Get(uniqueIndexValue)
			if existingUniqueVal != nil {
				return fmt.Errorf("can't add record, this would break unique index: %s", uniqueIndexKeys)
			}
		}

		//We increment the ID, the ID is used to avoid long seeking times
		nextId, _ := bucket.NextSequence()
		item.SetId(uint32(nextId))

		//We serialize the ID
		idBytes, err := toBytes(item.Id(), b.marshaler)
		if err != nil {
			return err
		}
		//Save the pkIndex for the id of the result.
		err = pkIndex.Add(indexValue, idBytes)
		if err != nil {
			return err
		}

		//Save the actual result, using the ID, not the key. The key is indexed so you can easily look up the ID
		serializedValue, err := b.marshaler.Marshal(item)
		if err != nil {
			return err
		}
		err = bucket.Put(idBytes, serializedValue)
		if err != nil {
			return err
		}
		if uniqueIndex != nil {
			err = uniqueIndex.Add(uniqueIndexValue, idBytes)
		}
		return err
	})
}

func (b *BoltStorage) Size() int64 {
	var count *int
	count = new(int)
	*count = 0
	_ = b.Database.View(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		stats := bucket.Stats()
		count = &stats.InlineBucketN
		return nil
	})
	return int64(*count)
}

//ForEach Goes through all the records
func (b *BoltStorage) ForEach(callback func(record interface{})) {
	_ = b.Database.View(func(tx *bolt.Tx) error {
		bucket := b.GetBucket(tx, resultsBucket)
		cursor := ReversibleCursor{C: bucket.Cursor(), Reverse: false}
		for _, val := cursor.First(); cursor.CanContinue(val); _, val = cursor.Next() {
			result, err := b.marshaler.Unmarshal(val)
			if err != nil {
				return err
			}
			callback(result)
		}
		return nil
	})
}

func (b *BoltStorage) GetNewest(count int) []search.ExternalResultItem {
	var output []search.ExternalResultItem
	_ = b.Database.View(func(tx *bolt.Tx) error {
		bucket := b.GetBucket(tx, resultsBucket)
		cursor := ReversibleCursor{C: bucket.Cursor(), Reverse: true}
		itemsFetched := 0
		for _, val := cursor.First(); cursor.CanContinue(val); _, val = cursor.Next() {
			if itemsFetched == count {
				break
			}
			newItem := search.ExternalResultItem{}
			if err := b.marshaler.UnmarshalAt(val, &newItem); err != nil {
				log.Warning("Couldn't deserialize item from bolt storage.")
				continue
			}
			output = append(output, newItem)
			itemsFetched++
		}
		return nil
	})
	return output
}

//StoreChat stores a new chat.
//The chat id is used as a keyParts.
//func (b *BoltStorage) StoreChat(chat *Chat) error {
//	//	defer db.Close()
//	err := b.Database.Update(func(tx *bolt.Tx) error {
//		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
//		if err != nil {
//			return err
//		}
//		key := i64tob(chat.ChatId)
//		val, err := b.marshaler.Marshal(chat)
//		if err != nil {
//			return err
//		}
//		return bucket.Put(key, val)
//	})
//	return err
//}

func DefaultBoltPath() string {
	cwd, _ := os.Getwd()
	return path.Join(cwd, "db", "bolt.db")
}

//createBucketIfItDoesntExist creates a new bucket by it's name if it doesn't exist
func (b *BoltStorage) createBucketIfItDoesntExist(tx *bolt.Tx, name string) (*bolt.Bucket, error) {
	if tx == nil || !tx.Writable() {
		return nil, errors.New("transaction is nil or not writable")
	}
	if name == "" {
		return nil, errors.New("bucket name is required")
	}
	var bucket *bolt.Bucket
	var err error
	bucketNames := append(b.rootBucket, name)
	//Make sure we keep our bucket structure correct.
	for _, bucketName := range bucketNames {
		if bucket != nil {
			if bucket, err = bucket.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return nil, err
			}
		} else {
			if bucket, err = tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return nil, err
			}
		}
	}
	return bucket, nil
}

// GetBucket returns the given bucket. You can use an array of strings for sub-buckets.
func (b *BoltStorage) GetBucket(tx *bolt.Tx, children ...string) *bolt.Bucket {
	var bucket *bolt.Bucket
	bucketNames := append(b.rootBucket, children...)
	for _, bucketName := range bucketNames {
		if bucket != nil {
			if bucket = bucket.Bucket([]byte(bucketName)); b == nil {
				return nil
			}
		} else {
			if bucket = tx.Bucket([]byte(bucketName)); b == nil {
				return nil
			}
		}
	}
	return bucket
}

//ForChat calls the callback for each chat, in an async way.
//func (b *BoltStorage) ForChat(callback func(chat *Chat)) error {
//	return b.Database.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("telegram_chats"))
//		if bucket == nil {
//			return nil
//		}
//		return bucket.ForEach(func(k, v []byte) error {
//			var chat = Chat{}
//			if err := b.marshaler.Unmarshal(v, &chat); err != nil {
//				return err
//			}
//			callback(&chat)
//			return nil
//		})
//	})
//}

//func (b *BoltStorage) GetChat(id int) (*Chat, error) {
//	var chat = Chat{}
//	found := false
//	err := b.Database.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("telegram_chats"))
//		if bucket == nil {
//			return nil
//		}
//		buff := bucket.Get(itob(id))
//		if buff == nil {
//			return nil
//		}
//		found = true
//		if err := b.marshaler.UnmarshalAt(buff, &chat); err != nil {
//			return err
//		}
//		return nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	if !found {
//		return nil, nil
//	}
//	return &chat, nil
//}

func (b *BoltStorage) Truncate() error {
	db := b.Database
	return db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}

//GetSearchResults by a given category id
func (b *BoltStorage) GetSearchResults(categoryId int) ([]search.ExternalResultItem, error) {
	bdb := b.Database
	var items []search.ExternalResultItem
	err := bdb.View(func(tx *bolt.Tx) error {
		var catName string
		if _, ok := categories.AllCategories[categoryId]; !ok {
			catName = "uncategorized"
		} else {
			catName = categories.AllCategories[categoryId].Name
		}

		bucket := tx.Bucket([]byte("searchResults")).Bucket([]byte(catName))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			var newItem search.ExternalResultItem
			err := b.marshaler.UnmarshalAt(v, &newItem)
			if err != nil {
				return err
			}
			items = append(items, newItem)
			return nil
		})
	})
	return items, err
}

//StoreSearchResults stores the given results
func (b *BoltStorage) StoreSearchResults(items []search.ExternalResultItem) error {
	db := b.Database
	for ix, item := range items {
		//the function passed to Batch may be called multiple times,
		err := db.Batch(func(tx *bolt.Tx) error {
			cgry := categories.AllCategories[item.Category]
			var cgryKey []byte
			if cgry == nil {
				cgryKey = []byte("uncategorized")
			} else {
				cgryKey = []byte(cgry.Name)
			}
			//Use the category as a keyParts
			bucket, _ := tx.CreateBucketIfNotExists([]byte("searchResults"))
			bucket, _ = bucket.CreateBucketIfNotExists(cgryKey)
			key, err := GetItemKey(item)
			if err != nil {
				return err
			}
			nextId, _ := bucket.NextSequence()
			item.ID = uint32(nextId)
			buf, err := b.marshaler.Marshal(item)
			if err != nil {
				return err
			}
			item.CreatedAt = time.Now()
			err = bucket.Put(key, buf)
			if err != nil {
				item.ID = 0
				log.Error(fmt.Sprintf("Error while inserting %d-th item. %s", ix, err))
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BoltStorage) SetNamespace(namespace string) {
	b.rootBucket = []string{namespace}
}

func GetItemKey(item search.ExternalResultItem) ([]byte, error) {
	if item.UUIDValue == "" {
		return nil, errors.New("record has no keyParts")
	}
	return []byte(item.UUIDValue), nil
}

// itob returns an 8-byte big endian representation of v.
//func uitob(v uint) []byte {
//	b := make([]byte, 8)
//	binary.BigEndian.PutUint64(b, uint64(v))
//	return b
//}
func i64tob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

//toBytes is a helper function that converts any value to bytes
func toBytes(key interface{}, codec *serializers.DynamicMarshaler) ([]byte, error) {
	if key == nil {
		return nil, nil
	}
	switch t := key.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	case int:
		return numbertob(int64(t))
	case uint:
		return numbertob(uint64(t))
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return numbertob(t)
	default:
		return codec.Marshal(key)
	}
}

func numbertob(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
