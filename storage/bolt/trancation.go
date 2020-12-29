package bolt

import "github.com/boltdb/bolt"

// Truncate the whole database
func (b *BoltStorage) Truncate() error {
	db := b.Database
	return db.Update(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return tx.DeleteBucket(name)
		})
	})
}
