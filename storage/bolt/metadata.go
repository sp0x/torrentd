package bolt

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
)

func (b *BoltStorage) setupMetadata() error {
	return b.Database.Update(func(tx *bolt.Tx) error {
		bucket, err := b.createBucketIfItDoesntExist(tx, resultsBucket)
		if err != nil {
			return err
		}
		b.loadMetadata(bucket)
		return nil
	})
}

func (b *BoltStorage) saveMetadata(bucket *bolt.Bucket) {
	metadataBytes, _ := json.Marshal(b.metadata)
	_ = bucket.Put([]byte("__meta"), metadataBytes)
}

func (b *BoltStorage) loadMetadata(bucket *bolt.Bucket) {
	metadataBytes := bucket.Get([]byte("__meta"))
	metadata := &Metadata{}
	if metadataBytes != nil {
		err := json.Unmarshal(metadataBytes, &metadata)
		if err != nil {
			panic(fmt.Sprintf("couldn't load metadata: %v", err))
		}
	}
	b.metadata = metadata
}

type Metadata struct {
	Indexes map[string]IndexMetadata `json:"indexes"`
}

type IndexMetadata struct {
	Name     string `json:"name"`
	Unique   bool   `json:"unique"`
	Location string
}

func (m *Metadata) AddIndex(name string, location string, isUnique bool) bool {
	if m.Indexes == nil {
		m.Indexes = make(map[string]IndexMetadata)
	}
	existingIx, exists := m.Indexes[name]
	if exists && existingIx.Location == location && existingIx.Unique == isUnique {
		return false
	}
	m.Indexes[name] = IndexMetadata{
		Name:     name,
		Unique:   isUnique,
		Location: location,
	}
	return true
}
