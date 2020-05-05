package db

import (
	"encoding/binary"
	"encoding/json"
	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"os"
	"path"
)

func open() (*bolt.DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return nil
	}
	dbPath := path.Join(cwd, "db", "bolt.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	return db, err
}

func StoreSearchResult(item *search.ExternalResultItem) error {
	db, err := open()
	if err != nil {
		return nil
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("searchResults"))
		nextId, _ := b.NextSequence()
		item.ID = uint(nextId)
		buf, err := json.Marshal(item)
		if err != nil {
			return err
		}
		//Use the id as a key
		return b.Put(uitob(item.ID), buf)
	})

}

// itob returns an 8-byte big endian representation of v.
func uitob(v uint) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
