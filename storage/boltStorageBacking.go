package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"os"
	"path"
	"time"
)

type BoltStorage struct {
	Database   *bolt.DB
	rootBucket []string
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
		Database: dbx,
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
		//Create all of our categories
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

//
func (b *BoltStorage) Find(query Query, result *search.ExternalResultItem) error {
	if query == nil {
		return errors.New("query is required")
	}
	bucketName := "results"
	return b.Database.View(func(tx *bolt.Tx) error {
		bucket := b.GetBucket(tx, bucketName)
		//Ways to go about this:
		//scan the entire bucket and filter by the query - this may be too slow
		//
		//Todo: implement querying if we're not using primary keys.
		serializedValue := toby
	})
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

func (b *BoltStorage) Update(query Query, item *search.ExternalResultItem) {
	panic("implement me")
}

func (b *BoltStorage) Create(item *search.ExternalResultItem) {

	panic("implement me")
}

//StoreChat stores a new chat.
//The chat id is used as a key.
func (b *BoltStorage) StoreChat(chat *Chat) error {
	//	defer db.Close()
	err := b.Database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("telegram_chats"))
		key := i64tob(chat.ChatId)
		val, err := json.Marshal(chat)
		if err != nil {
			return err
		}
		return b.Put(key, val)
	})
	return err
}

func DefaultBoltPath() string {
	cwd, _ := os.Getwd()
	return path.Join(cwd, "db", "bolt.db")
}

//ForChat calls the callback for each chat, in an async way.
func (b *BoltStorage) ForChat(callback func(chat *Chat)) error {
	return b.Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("telegram_chats"))
		return b.ForEach(func(k, v []byte) error {
			var chat = Chat{}
			if err := json.Unmarshal(v, &chat); err != nil {
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
		b := tx.Bucket([]byte("telegram_chats"))
		buff := b.Get(itob(id))
		if buff == nil {
			return nil
		}
		found = true
		if err := json.Unmarshal(buff, &chat); err != nil {
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

		b := tx.Bucket([]byte("searchResults")).Bucket([]byte(catName))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var newItem search.ExternalResultItem
			err := json.Unmarshal(v, &newItem)
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
			//Use the category as a key
			b, _ := tx.CreateBucketIfNotExists([]byte("searchResults"))
			b, _ = b.CreateBucketIfNotExists(cgryKey)
			key, err := getItemKey(item)
			if err != nil {
				return err
			}
			nextId, _ := b.NextSequence()
			item.ID = uint(nextId)
			buf, err := json.Marshal(item)
			if err != nil {
				return err
			}
			item.CreatedAt = time.Now()
			err = b.Put(key, buf)
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
		return nil, errors.New("record has no key")
	}
	return []byte(item.GUID), nil
}

// itob returns an 8-byte big endian representation of v.
func uitob(v uint) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
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

func toBytes(key interface{}, codec codec.MarshalUnmarshaler) ([]byte, error) {
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
