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

type BoltStorage struct {
	Database   *bolt.DB
	rootBucket []string
	marshaler  serializers.MarshalUnmarshaler
}

func NewBoltStorage(dbPath string) (*BoltStorage, error) {
	if dbPath == "" {
		dbPath = DefaultBoltPath()
	}
	dbx, err := GetBoltDb(dbPath)
	if err != nil {
		return nil, err
	}
	bls := &BoltStorage{
		Database:  dbx,
		marshaler: json.Serializer,
	}
	return bls, nil
}

var categoriesInitialized = false

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
		_, err = tx.CreateBucketIfNotExists([]byte("telegram_chats"))
		if err != nil {
			return err
		}
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

type ChatMessage struct {
	Text   string
	ChatId string
}

type Chat struct {
	Username    string
	InitialText string
	ChatId      int64
}

const (
	resultsBucket = "results"
)

//Find something by it's index keys.
func (b *BoltStorage) Find(query indexing.Query, result *search.ExternalResultItem) error {
	if query == nil {
		return errors.New("query is required")
	}
	return b.Database.View(func(tx *bolt.Tx) error {
		bucket := b.GetBucket(tx, resultsBucket)
		if bucket == nil {
			return errors.New("not found")
		}
		//Ways to go about this:
		//-scan the entire bucket and filter by the query - this may be too slow
		//-serialize the keyParts query and use it as the keyParts in the bucket, to search by it
		//-use the query as an index, having a bucket with all the ids
		//convert the query to a prefix, and seek to the start of that prefix in a cursor
		//iterate until the end of the prefixed region in the cursor
		//Todo: implement querying if we're not using primary keys.
		idx, err := GetIndexFromQuery(bucket, query)
		if err != nil {
			return err
		}
		indexValue := indexing.GetIndexValueFromQuery(query)
		ids := idx.All(indexValue, indexing.SingleItemCursor())
		if len(ids) == 0 {
			return errors.New("not found")
		}
		rawResult := bucket.Get(ids[0])
		err = b.marshaler.Unmarshal(rawResult, result)
		if err != nil {
			return err
		}
		return nil
	})
}

func (b *BoltStorage) Update(query indexing.Query, item *search.ExternalResultItem) error {
	if query == nil {
		return errors.New("query is required")
	}
	return b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		idx, err := GetIndexFromQuery(bucket, query)
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
func (b *BoltStorage) Create(item *search.ExternalResultItem, additionalPK indexing.Key) error {
	item.GUID = uuid.New().String()
	key := indexing.NewKey("GUID")
	err := b.CreateWithId(key, item, nil)
	if err != nil {
		return err
	}
	//If we don't have an unique index, we can stop here.
	if len(additionalPK) == 0 {
		return nil
	}
	indexValue := indexing.GetIndexValueFromItem(additionalPK, item)
	//We need add a new index: additionalPK -> GUID
	return b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		//We get the keyIndex that we'll use
		keyToGuidIndex, err := GetUniqueIndexFromKeys(bucket, additionalPK)
		if err != nil {
			return err
		}
		guidBytes := []byte(item.GUID)
		//Save the keyIndex for the id of the result.
		err = keyToGuidIndex.Add(indexValue, guidBytes)
		return err
	})
}

//CreateWithId a new record for a result.
//The key is used if you have a custom object that uses a different key, not the GUID
func (b *BoltStorage) CreateWithId(keyParts indexing.Key, item *search.ExternalResultItem, uniqueIndexKeys indexing.Key) error {
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
		pkIndex, err := GetUniqueIndexFromKeys(bucket, keyParts)
		if err != nil {
			return err
		}
		var uniqueIndex indexing.Index
		if len(uniqueIndexKeys) > 0 {
			uniqueIndex, err = GetUniqueIndexFromKeys(bucket, uniqueIndexKeys)
			if err != nil {
				return err
			}
			existingUniqueVal := uniqueIndex.Get(uniqueIndexValue)
			if existingUniqueVal != nil {
				return fmt.Errorf("can't add record, this would break unique index: %s", uniqueIndexKeys)
			}
		}

		//We increment the ID
		nextId, _ := bucket.NextSequence()
		item.ID = uint32(nextId)

		//We serialize the ID
		idBytes, err := toBytes(item.ID, b.marshaler)
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

func (b *BoltStorage) GetNewest(count int) []search.ExternalResultItem {
	var output []search.ExternalResultItem
	_ = b.Database.View(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		cursor := ReversibleCursor{C: bucket.Cursor(), Reverse: true}
		itemsFetched := 0
		for _, val := cursor.First(); cursor.CanContinue(val); _, val = cursor.Next() {
			if itemsFetched == count {
				break
			}
			newItem := search.ExternalResultItem{}
			if err := b.marshaler.Unmarshal(val, &newItem); err != nil {
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
func (b *BoltStorage) StoreChat(chat *Chat) error {
	//	defer db.Close()
	err := b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, "telegram_chats")
		if err != nil {
			return err
		}
		key := i64tob(chat.ChatId)
		val, err := b.marshaler.Marshal(chat)
		if err != nil {
			return err
		}
		return bucket.Put(key, val)
	})
	return err
}

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
func (b *BoltStorage) ForChat(callback func(chat *Chat)) error {
	return b.Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("telegram_chats"))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			var chat = Chat{}
			if err := b.marshaler.Unmarshal(v, &chat); err != nil {
				return err
			}
			callback(&chat)
			return nil
		})
	})
}

func (b *BoltStorage) GetChat(id int) (*Chat, error) {
	var chat = Chat{}
	found := false
	err := b.Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("telegram_chats"))
		if bucket == nil {
			return nil
		}
		buff := bucket.Get(itob(id))
		if buff == nil {
			return nil
		}
		found = true
		if err := b.marshaler.Unmarshal(buff, &chat); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &chat, nil
}

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
			err := b.marshaler.Unmarshal(v, &newItem)
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
			key, err := getItemKey(item)
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

func getItemKey(item search.ExternalResultItem) ([]byte, error) {
	if item.GUID == "" {
		return nil, errors.New("record has no keyParts")
	}
	return []byte(item.GUID), nil
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
func toBytes(key interface{}, codec serializers.MarshalUnmarshaler) ([]byte, error) {
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
