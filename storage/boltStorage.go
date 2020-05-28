package storage

import (
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
	Database *bolt.DB
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
