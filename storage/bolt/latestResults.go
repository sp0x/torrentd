package bolt

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"strconv"
)

const (
	latestResultsBucket     = "results.latest"
	latestResultsIndexKey   = "__index"
	latestResultsBucketSize = 20
)

//updateLatestResults updates a bucket with the latest 20 results. the items are unordered
func (b *BoltStorage) updateLatestResults(tx *bolt.Tx, item search.Record) error {
	bucket, err := b.assertBucket(tx, latestResultsBucket)
	if err != nil {
		return err
	}
	latestKey := bucket.Get([]byte(latestResultsIndexKey))
	var nextKey []byte
	if latestKey == nil {
		nextKey = []byte("0")
	} else {
		indexInt, _ := strconv.Atoi(string(latestKey))
		if indexInt > latestResultsBucketSize {
			indexInt = -1
		}
		nextKey = []byte(fmt.Sprintf("%d", indexInt+1))
	}

	serializedValue, err := b.marshaler.Marshal(item)
	if err != nil {
		return err
	}
	err = bucket.Put(nextKey, serializedValue)
	if err != nil {
		return err
	}
	err = bucket.Put([]byte(latestResultsIndexKey), nextKey)
	return err
}

func (b *BoltStorage) getLatestResultsCursor(tx *bolt.Tx) (indexing.Cursor, error) {
	bucket := b.GetRootBucket(tx, latestResultsBucket)
	cursor := bucket.Cursor()
	return &FilteredCursor{C: cursor, Filters: []func([]byte, []byte) bool{
		func(id []byte, value []byte) bool {
			return string(id) == latestResultsIndexKey
		},
	}}, nil
}
