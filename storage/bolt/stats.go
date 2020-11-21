package bolt

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/sp0x/torrentd/storage/stats"
	"time"
)

func (b *BoltStorage) GetStats(dumpDb bool) *stats.Stats {
	if dumpDb {
		b.dumpDbBuckets()
	}
	//Go over each namespace
	output := &stats.Stats{}
	_ = b.Database.View(func(tx *bolt.Tx) error {
		for _, ns := range b.getNamespaces(tx) {
			size := getNamespaceSize(tx, b, ns)
			nsStats := stats.NamespaceStats{
				Name:        ns,
				RecordCount: size,
				LastUpdated: time.Time{},
			}
			output.Namespaces = append(output.Namespaces, nsStats)
		}
		return nil
	})
	return output
}

func (b *BoltStorage) dumpDbBuckets() {
	_ = b.Database.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			bucketStats := bucket.Stats()
			_, _ = fmt.Printf("bucket `%s`:\t%d keys, depth: %d, buckets: %d\n", name, bucketStats.KeyN, bucketStats.Depth, bucketStats.BucketN)
			dumpBucket(bucket)
			return nil
		})
	})
}

func dumpBucket(bucket *bolt.Bucket) {
	_ = bucket.ForEach(func(key []byte, val []byte) error {
		subBucket := bucket.Bucket(key)
		_, _ = fmt.Printf("  %s:\t%s", key, val)
		if subBucket != nil {
			bucketStats := subBucket.Stats()
			_, _ = fmt.Printf("%d keys, depth: %d, buckets: %d", bucketStats.KeyN, bucketStats.Depth, bucketStats.BucketN)
		}
		fmt.Print("\n")
		return nil
	})
}

func getNamespaceSize(tx *bolt.Tx, b *BoltStorage, name string) int {
	bucket := b.GetRootBucket(tx, name, namespaceResultsBucketName)
	if bucket == nil {
		return 0
	}
	bucketStats := bucket.Stats()
	return bucketStats.KeyN
}

//Size gets the record count in this namespace's results bucket
func (b *BoltStorage) Size() int64 {
	var count *int
	count = new(int)
	*count = 0
	_ = b.Database.View(func(tx *bolt.Tx) error {
		nsz := getNamespaceSize(tx, b, b.rootBucket[0])
		count = &nsz
		return nil
	})
	return int64(*count)
}

func (b *BoltStorage) getNamespaces(tx *bolt.Tx) []string {
	var names []string
	_ = tx.ForEach(func(name []byte, b *bolt.Bucket) error {
		nameStr := string(name)
		if nameStr != internalBucketName && nameStr != categoriesBucketName && nameStr != latestResultsBucketName {
			names = append(names, nameStr)
		}
		return nil
	})
	return names
}
