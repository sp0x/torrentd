package bolt

import (
	"errors"
	"strings"

	"github.com/boltdb/bolt"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
)

const (
	indexPrefix = "__index_"
)

// getIndexFromQuery gets the index from the fields in a query
func (b *BoltStorage) GetIndexFromQuery(bucket *bolt.Bucket, query indexing.Query) (indexing.Index, error) {
	indexName := indexing.GetIndexNameFromQuery(query)
	return b.getIndex(bucket, "unique", indexName)
}

// GetUniqueIndexFromKeys gets the index from a key.
func (b *BoltStorage) GetUniqueIndexFromKeys(bucket *bolt.Bucket, keyParts *indexing.Key) (indexing.Index, error) {
	indexName := strings.Join(keyParts.Fields, "_")
	return b.getIndex(bucket, "unique", indexName)
}

// hasIndex Figures out if an index exists.
func hasIndex(bucket *bolt.Bucket, name string) bool {
	indexName := []byte(indexPrefix + name)
	val := bucket.Bucket(indexName)
	return val != nil
}

// getIndex creates a new index if one doesn't exist for a bucket
func (b *BoltStorage) getIndex(bucket *bolt.Bucket, kind string, name string) (indexing.Index, error) {
	var index indexing.Index
	var err error
	indexName := []byte(indexPrefix + name)
	switch kind {
	case "unique":
		index, err = NewUniqueIndex(bucket, indexName)
	case "id":
		index, err = NewListIndex(bucket, indexName)
	}
	if err != nil {
		return nil, err
	}

	// Add the information for this indexer.
	if b.metadata.AddIndex(name, string(indexName), kind == "unique") {
		b.saveMetadata(bucket)
	}
	return index, err
}

func getByIndex(bucket *bolt.Bucket, index indexing.Index, indexValue []byte) (id []byte, result []byte, err error) {
	id = indexing.First(index, indexValue)
	if id == nil {
		return nil, nil, errors.New("not found")
	}
	// Use the same root bucket and query by the ID in our index
	result = bucket.Get(id)
	return id, result, nil
}

func getDefaultPK() *indexing.Key {
	return indexing.NewKey("UUID")
}

func getDefaultPKIndex(b *BoltStorage, bucket *bolt.Bucket) (indexing.Index, error) {
	return b.GetUniqueIndexFromKeys(bucket, getDefaultPK())
}

func getByDefaultPKIndex(b *BoltStorage, bucket *bolt.Bucket, id []byte) ([]byte, error) {
	defaultPKIndex, err := getDefaultPKIndex(b, bucket)
	if err != nil {
		return nil, err
	}
	ids := defaultPKIndex.Get(id)
	rawResult := bucket.Get(ids)
	return rawResult, nil
}

func getRecordByIndexOrDefault(b *BoltStorage, bucket *bolt.Bucket, dbIndex indexing.Index, indexValue []byte, result interface{}) error {
	id, rawResult, err := getByIndex(bucket, dbIndex, indexValue)
	if err != nil {
		return err
	}
	// If the result is nil, we might be using the additional PK
	if rawResult == nil {
		// TODO: add the pk information into the root bucket, so we know if we need to do this.
		rawResult, err = getByDefaultPKIndex(b, bucket, id)
	}
	if err != nil {
		return err
	}
	err = b.marshaler.UnmarshalAt(rawResult, result)
	return err
}

func (b *BoltStorage) getFromIndexedQuery(bucketName string, tx *bolt.Tx, query indexing.Query, result interface{}) error {
	indexValue := indexing.GetIndexValueFromQuery(query)
	bucket := b.GetBucket(tx, bucketName)
	if bucket == nil {
		return errors.New("not found")
	}
	idx, err := b.GetIndexFromQuery(bucket, query)
	if err != nil {
		return err
	}
	return getRecordByIndexOrDefault(b, bucket, idx, indexValue, result)
}

func (b *BoltStorage) indexQuery(bucketName string, query indexing.Query) error {
	return b.Database.Update(func(tx *bolt.Tx) error {
		_, err := b.GetIndexFromQuery(b.GetBucket(tx, bucketName), query)
		return err
	})
}

func GetPKValueFromRecord(item search.Record) ([]byte, error) {
	if item.UUID() == "" {
		return nil, errors.New("record has no keyParts")
	}
	return []byte(item.UUID()), nil
}
