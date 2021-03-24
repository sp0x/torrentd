package bolt

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/boltdb/bolt"

	"github.com/sp0x/torrentd/storage/indexing"
)

type Metadata struct {
	Indexes map[string]indexing.IndexMetadata `json:"indexes"`
}

func (b *Storage) isNamespaced() bool {
	return b.rootBucket != nil
}

func (b *Storage) setupMetadata() error {
	return b.Database.Update(func(tx *bolt.Tx) error {
		internalBucket, err := b.assertBucket(tx, internalBucketName)
		if err != nil {
			return err
		}
		if b.isNamespaced() {
			nsBucket, err := b.assertNamespaceBucket(tx, namespaceResultsBucketName)
			if err != nil {
				return err
			}
			b.loadMetadata(nsBucket)
		}
		b.loadGlobalMetadata(internalBucket)
		return nil
	})
}

func (b *Storage) saveMetadata(bucket *bolt.Bucket) {
	metadataBytes, _ := json.Marshal(b.metadata)
	_ = bucket.Put([]byte(metaBucketName), metadataBytes)
}

func (b *Storage) loadMetadata(bucket *bolt.Bucket) {
	metadataBytes := bucket.Get([]byte(metaBucketName))
	metadata := &Metadata{}
	if metadataBytes != nil {
		err := json.Unmarshal(metadataBytes, metadata)
		if err != nil {
			fmt.Printf("couldn't load metadata: %v", err)
			os.Exit(1)
		}
	}
	b.metadata = metadata
	// if b.isNamespaced() {
	//	b.metadata = metadata
	// }
}

func (b *Storage) GetIndexes() map[string]indexing.IndexMetadata {
	if b.metadata == nil {
		return nil
	}
	return b.metadata.Indexes
}

func (b *Storage) HasIndex(meta *indexing.IndexMetadata) bool {
	_, found := b.metadata.Indexes[meta.Name]
	return found
}

func (b *Storage) HasIndexWithName(name string) bool {
	_, found := b.metadata.Indexes[name]
	return found
}

func (m *Metadata) AddIndex(name string, location string, isUnique bool) bool {
	if m.Indexes == nil {
		m.Indexes = make(map[string]indexing.IndexMetadata)
	}
	existingIx, exists := m.Indexes[name]
	if exists && existingIx.Location == location && existingIx.Unique == isUnique {
		return false
	}
	m.Indexes[name] = indexing.IndexMetadata{
		Name:     name,
		Unique:   isUnique,
		Location: location,
	}
	return true
}
