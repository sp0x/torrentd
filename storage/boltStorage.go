package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/categories"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"os"
	"path"
)

type BoltStorage struct{}

var categoriesInitialized = false

func GetBoltDb() (*bolt.DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	dbPath := path.Join(cwd, "db", "bolt.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}
	//Setup our DB
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("searchResults"))
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
	return db, nil
}

//StoreSearchResults stores the given results
func (b *BoltStorage) StoreSearchResults(items []search.ExternalResultItem) error {
	db, err := GetBoltDb()
	if err != nil {
		return nil
	}
	defer func() {
		_ = db.Close()
	}()
	for ix, item := range items {

		//the function passed to Batch may be called multiple times,
		err = db.Batch(func(tx *bolt.Tx) error {
			cgry := categories.AllCategories[item.Category]
			var cgryKey []byte
			if cgry == nil {
				cgryKey = []byte("uncategorized")
			} else {
				cgryKey = []byte(cgry.Name)
			}
			//Use the category as a key
			b := tx.Bucket([]byte("searchResults")).Bucket(cgryKey)
			// Find last key in bucket, decode as bigendian uint64, increment
			// by one, encode back to []byte, and add new key.
			nextId, _ := b.NextSequence()
			//Use the id as a key
			key, err := getItemKey(item)
			if err != nil {
				return err
			}
			item.ID = uint(nextId)
			buf, err := json.Marshal(item)
			if err != nil {
				return err
			}
			err = b.Put(key, buf)
			if err != nil {
				item.ID = 0
				log.Error("Error while inserting %s-th item. %s", ix, err)
				return err
			}
			return nil
		})
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
