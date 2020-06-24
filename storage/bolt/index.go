package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/sp0x/torrentd/storage/indexing"
	"strings"
)

const (
	indexPrefix = "__index_"
)

//getIndexFromQuery gets the index from the fields in a query
func (b *BoltStorage) GetIndexFromQuery(bucket *bolt.Bucket, query indexing.Query) (indexing.Index, error) {
	indexName := indexing.GetIndexNameFromQuery(query)
	return b.getIndex(bucket, "unique", indexName)
}

//GetUniqueIndexFromKeys gets the index from a key.
func (b *BoltStorage) GetUniqueIndexFromKeys(bucket *bolt.Bucket, keyParts *indexing.Key) (indexing.Index, error) {
	indexName := strings.Join(keyParts.Fields, "_")
	return b.getIndex(bucket, "unique", indexName)
}

//hasIndex Figures out if an index exists.
func hasIndex(bucket *bolt.Bucket, name string) bool {
	indexName := []byte(indexPrefix + name)
	val := bucket.Bucket(indexName)
	return val != nil
}

//getIndex creates a new index if one doesn't exist for a bucket
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

	//Add the information for this indexer.
	if b.metadata.AddIndex(name, string(indexName), kind == "unique") {
		b.saveMetadata(bucket)
	}
	return index, err
}
