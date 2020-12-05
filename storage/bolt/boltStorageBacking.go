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
	"reflect"
	"time"
)

/**
Storage scheme:
bucket:
 - indexes(unique/non): bucket of IDs
 - items
 - __meta: indexes information
*/
const (
	internalBucketName         = "__internal"
	namespaceResultsBucketName = "results"
	metaBucketName             = "__meta"
	categoriesBucketName       = "__categories"
)

var categoriesInitialized = false

type BoltStorage struct {
	Database   *bolt.DB
	rootBucket []string
	marshaler  *serializers.DynamicMarshaler
	metadata   *Metadata
	recordType reflect.Type
}

func ensurePathExists(dbPath string) {
	if dbPath == "" {
		return
	}

	dirPath := path.Dir(dbPath)
	_ = os.MkdirAll(dirPath, os.ModePerm)
}

//NewBoltStorage - opens a BoltDB storage file
func NewBoltStorage(dbPath string, recordTypePtr interface{}) (*BoltStorage, error) {
	if dbPath == "" {
		dbPath = DefaultBoltPath()
	}
	if reflect.TypeOf(recordTypePtr).Kind() != reflect.Ptr {
		return nil, errors.New("recordTypePtr must be a pointer type")
	}
	ensurePathExists(dbPath)
	dbx, err := GetBoltDb(dbPath)
	if err != nil {
		return nil, err
	}
	bolts := &BoltStorage{
		Database:   dbx,
		marshaler:  serializers.NewDynamicMarshaler(recordTypePtr, json.Serializer),
		recordType: reflect.Indirect(reflect.ValueOf(recordTypePtr)).Type(),
	}
	err = bolts.setupMetadata()
	if err != nil {
		bolts.Close()
		return nil, err
	}
	return bolts, nil
}

func GetBoltDb(file string) (*bolt.DB, error) {
	dbx, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = setupCategories(dbx)
	if err != nil {
		return nil, err
	}
	return dbx, nil
}

func setupCategories(db *bolt.DB) error {
	//Setup our DB
	err := db.Update(func(tx *bolt.Tx) error {
		ctgBucket, err := tx.CreateBucketIfNotExists([]byte(categoriesBucketName))
		if err != nil {
			return err
		}

		//CreateWithId all of our categories
		if !categoriesInitialized {
			for _, cat := range categories.AllCategories {
				catKey := []byte(cat.Name)
				_, err := ctgBucket.CreateBucketIfNotExists(catKey)
				if err != nil {
					return err
				}
			}
		}
		categoriesInitialized = true
		return err
	})
	return err
}

func (b *BoltStorage) Close() {
	if b.Database == nil {
		return
	}
	_ = b.Database.Close()
}

//Find records by it's index keys.
func (b *BoltStorage) Find(query indexing.Query, result interface{}) error {
	if query == nil {
		return errors.New("query is required")
	}
	//The our bucket, and the index that matches the query best
	err := b.Database.View(func(tx *bolt.Tx) error {
		return b.getFromIndexedQuery(namespaceResultsBucketName, tx, query, result)
	})
	//At this point we can quit.
	if err == nil {
		return nil
	}
	//If the index does not exist, we create it and query by it
	if _, ok := err.(*IndexDoesNotExistAndNotWritable); ok {
		err = b.indexQuery(namespaceResultsBucketName, query)
		if err != nil {
			return err
		}
		err = b.Database.View(func(tx *bolt.Tx) error {
			return b.getFromIndexedQuery(namespaceResultsBucketName, tx, query, result)
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
		bucket, err := b.assertNamespaceBucket(tx, namespaceResultsBucketName)
		if err != nil {
			return err
		}
		queryIndex, err := b.GetIndexFromQuery(bucket, query)
		if err != nil {
			return err
		}
		indexValue := indexing.GetIndexValueFromQuery(query)
		//Fetch the ID from the index
		ids := queryIndex.All(indexValue, indexing.SingleItemCursor())
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
	err := b.CreateWithId(getDefaultPK(), item, nil)
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
		bucket, err := b.assertNamespaceBucket(tx, namespaceResultsBucketName)
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
		bucket, err := b.assertNamespaceBucket(tx, namespaceResultsBucketName)
		if err != nil {
			return err
		}
		//We get the primaryIndex that we'll use
		primaryIndex, err := b.GetUniqueIndexFromKeys(bucket, keyParts)
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
		//Save the primaryIndex for the id of the result.
		err = primaryIndex.Add(indexValue, idBytes)
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
		if err != nil {
			return nil
		}

		return b.PushToLatestItems(tx, serializedValue)
	})
}

//ForEach Goes through all the records
func (b *BoltStorage) ForEach(callback func(record search.ResultItemBase)) {
	_ = b.Database.View(func(tx *bolt.Tx) error {
		bucket := b.GetBucket(tx, namespaceResultsBucketName)
		cursor := ReversibleCursor{C: bucket.Cursor(), Reverse: false}
		for _, val := cursor.First(); cursor.CanContinue(val); _, val = cursor.Next() {
			result, err := b.marshaler.Unmarshal(val)
			if err != nil {
				return err
			}
			callback(result.(search.ResultItemBase))
		}
		return nil
	})
}

func DefaultBoltPath() string {
	cwd, _ := os.Getwd()
	return path.Join(cwd, "db", "bolt.db")
}

//assertBucket makes sure a bucket exists, in the given path
func (b *BoltStorage) assertBucket(tx *bolt.Tx, fullName ...string) (*bolt.Bucket, error) {
	if tx == nil || !tx.Writable() {
		return nil, errors.New("transaction is nil or not writable")
	}
	if fullName == nil {
		return nil, errors.New("bucket name is required")
	}
	var bucket *bolt.Bucket
	var err error
	//Make sure we keep our bucket structure correct.
	for _, bucketName := range fullName {
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

//assertNamespaceBucket creates a new bucket by it's name if it doesn't exist, in the preset namespace
func (b *BoltStorage) assertNamespaceBucket(tx *bolt.Tx, name string) (*bolt.Bucket, error) {
	if tx == nil || !tx.Writable() {
		return nil, errors.New("transaction is nil or not writable")
	}
	if name == "" {
		return nil, errors.New("bucket name is required")
	}
	bucketNames := append(b.rootBucket, name)
	return b.assertBucket(tx, bucketNames...)
}

// GetBucket returns the given bucket. You can use an array of strings for sub-buckets.
func (b *BoltStorage) GetBucket(tx *bolt.Tx, children ...string) *bolt.Bucket {
	bucketNamespace := append(b.rootBucket, children...)
	return b.GetRootBucket(tx, bucketNamespace...)
}

func (b *BoltStorage) GetRootBucket(tx *bolt.Tx, children ...string) *bolt.Bucket {
	var bucket *bolt.Bucket
	bucketNamespace := children
	for _, bucketName := range bucketNamespace {
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

//GetSearchResults by a given category id
func (b *BoltStorage) GetSearchResults(categoryId int) ([]search.ScrapeResultItem, error) {
	bdb := b.Database
	var items []search.ScrapeResultItem
	err := bdb.View(func(tx *bolt.Tx) error {
		var catName string
		if _, ok := categories.AllCategories[categoryId]; !ok {
			catName = "uncategorized"
		} else {
			catName = categories.AllCategories[categoryId].Name
		}

		categoryBucket := tx.Bucket([]byte(categoriesBucketName)).Bucket([]byte(catName))
		if categoryBucket == nil {
			return nil
		}
		return categoryBucket.ForEach(func(k, v []byte) error {
			var newItem search.ScrapeResultItem
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
func (b *BoltStorage) StoreSearchResults(items []search.ScrapeResultItem) error {
	db := b.Database
	for ix, item := range items {
		//the function passed to Batch may be called multiple times,
		err := db.Batch(func(tx *bolt.Tx) error {
			categoryObj := item.GetFieldWithDefault("category", categories.Uncategorized).(*categories.Category)
			cgryKey := []byte(categoryObj.Name)
			//Use the category as a keyParts
			categoriesBucket, _ := tx.CreateBucketIfNotExists([]byte(categoriesBucketName))
			categoriesBucket, _ = categoriesBucket.CreateBucketIfNotExists(cgryKey)
			key, err := GetPKValueFromRecord(&item)
			if err != nil {
				return err
			}
			nextId, _ := categoriesBucket.NextSequence()
			item.ID = uint32(nextId)
			buf, err := b.marshaler.Marshal(item)
			if err != nil {
				return err
			}
			item.CreatedAt = time.Now()
			err = categoriesBucket.Put(key, buf)
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

//Set the root namespace
func (b *BoltStorage) SetNamespace(namespace string) error {
	b.rootBucket = []string{namespace}
	err := b.setupMetadata()
	if err != nil {
		fmt.Printf("Couldn't set namespace `%s`, failed while setting up meta-data: %v", namespace, err)
		return err
	}
	return err
}

func (b *BoltStorage) loadGlobalMetadata(bucket *bolt.Bucket) {

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
