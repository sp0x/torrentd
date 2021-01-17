package bolt

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
)

const (
	latestResultsBucketName = "results.latest"
	latestResultsIndexKey   = "__index"
	latestResultsBucketSize = 20
)

// PushToLatestItems updates a bucket with the latest 20 results. the items are unordered
func (b *Storage) PushToLatestItems(tx *bolt.Tx, serializedItem []byte) error {
	if serializedItem == nil {
		return errors.New("serialized value is required")
	}
	bucket, err := b.assertBucket(tx, latestResultsBucketName)
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
	err = bucket.Put(nextKey, serializedItem)
	if err != nil {
		return err
	}
	err = bucket.Put([]byte(latestResultsIndexKey), nextKey)
	return err
}

func (b *Storage) getLatestResultsCursor(tx *bolt.Tx) (indexing.Cursor, error) {
	bucket := b.GetRootBucket(tx, latestResultsBucketName)
	if bucket == nil {
		return nil, errors.New("root bucket doesn't exist")
	}
	cursor := bucket.Cursor()
	return &FilteredCursor{C: cursor, Filters: []func([]byte, []byte) bool{
		func(id []byte, value []byte) bool {
			return string(id) == latestResultsIndexKey
		},
	}}, nil
}

// GetLatest gets the newest results for all the indexes
func (b *Storage) GetLatest(count int) []search.ResultItemBase {
	var output []search.ResultItemBase
	_ = b.Database.View(func(tx *bolt.Tx) error {
		cursor, err := b.getLatestResultsCursor(tx)
		if err != nil {
			return err
		}
		itemsFetched := 0
		for _, value := cursor.First(); value != nil && cursor.CanContinue(value); _, value = cursor.Next() {
			newItem := reflect.New(b.recordType).Interface().(search.ResultItemBase)
			if err := b.marshaler.UnmarshalAt(value, &newItem); err != nil {
				log.Warning("Couldn't deserialize item from bolt storage.")
				break
			}
			output = append(output, newItem)
			itemsFetched++
			if itemsFetched == count {
				break
			}
		}
		return nil
	})
	return output
}
