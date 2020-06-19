package storage

import (
	"github.com/boltdb/bolt"
	"strings"
)

const (
	indexPrefix = "__index_"
)

//getIndexFrommQuery gets the index from the fields in a query
func getIndexFromQuery(bucket *bolt.Bucket, query Query) (Index, error) {
	indexName := GetIndexNameFromQuery(query)
	return getIndex(bucket, "unique", indexName)
}

//getIndexFromKeys gets the index from a key.
func getIndexFromKeys(bucket *bolt.Bucket, keyParts Key) (Index, error) {
	indexName := strings.Join(keyParts, "_")
	return getIndex(bucket, "unique", indexName)
}

//getIndex creates a new index if one doesn't exist for a bucket
func getIndex(bucket *bolt.Bucket, kind string, name string) (Index, error) {
	var index Index
	var err error
	indexName := []byte(indexPrefix + name)
	switch kind {
	case "unique":
		index, err = NewUniqueIndex(bucket, indexName)
	case "id":
		index, err = NewListIndex(bucket, indexName)
	}
	return index, err
}

type Index interface {
	Add(value []byte, targetID []byte) error
	Remove(value []byte) error
	RemoveById(id []byte) error
	Get(value []byte) []byte
	All(value []byte, opts *CursorOptions) [][]byte
	AllRecords(opts *CursorOptions) [][]byte
	Range(min []byte, max []byte, opts *CursorOptions) [][]byte
	//	AllWithPrefix(prefix []byte, opts *CursorOptions) ([][]byte)
}
