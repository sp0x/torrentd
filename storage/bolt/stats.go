package bolt

import "github.com/boltdb/bolt"

//Size gets the record count in this namespace's results bucket
func (b *BoltStorage) Size() int64 {
	var count *int
	count = new(int)
	*count = 0
	_ = b.Database.View(func(tx *bolt.Tx) error {
		bucket, err := b.assertNamespaceBucket(tx, resultsBucket)
		if err != nil {
			return err
		}
		stats := bucket.Stats()
		count = &stats.InlineBucketN
		return nil
	})
	return int64(*count)
}
