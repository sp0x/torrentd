package bolt

import (
	"bytes"
	"errors"

	"github.com/boltdb/bolt"

	"github.com/sp0x/torrentd/storage/indexing"
)

type IndexDoesNotExistAndNotWritable struct{}

func (e *IndexDoesNotExistAndNotWritable) Error() string {
	return "index does not exist and couln't be created"
}

// NewUniqueIndex creates a new unique index bucket
func NewUniqueIndex(parentBucket *bolt.Bucket, name []byte) (*UniqueIndex, error) {
	var err error
	if parentBucket == nil {
		return nil, errors.New("parent bucket is required")
	}
	bucket := parentBucket.Bucket(name)
	if bucket == nil {
		if !parentBucket.Writable() {
			return nil, &IndexDoesNotExistAndNotWritable{}
		}
		bucket, err = parentBucket.CreateBucket(name)
		if err != nil {
			return nil, err
		}
	}
	return &UniqueIndex{
		IndexBucket:  bucket,
		ParentBucket: parentBucket,
	}, nil
}

// UniqueIndex ensures the indexed values are all unique.
type UniqueIndex struct {
	ParentBucket *bolt.Bucket
	IndexBucket  *bolt.Bucket
}

// Add a value to the unique index. We're using the index value as a keyParts
// and we're storing the id in there.
func (ix *UniqueIndex) Add(indexValue []byte, id []byte) error {
	if indexValue == nil {
		return errors.New("indexValue is required")
	}
	if id == nil {
		return errors.New("id is required")
	}
	existingIndex := ix.IndexBucket.Get(indexValue)
	if existingIndex != nil {
		if bytes.Equal(existingIndex, id) {
			return nil
		}
		return errors.New("index indexValue already exists")
	}
	return ix.IndexBucket.Put(indexValue, id)
}

// Remove a index value from the unique index.
func (ix *UniqueIndex) Remove(indexValue []byte) error {
	return ix.IndexBucket.Delete(indexValue)
}

// RemoveByID removes the first id from the index that matches the given id.
func (ix *UniqueIndex) RemoveByID(id []byte) error {
	cursor := ix.IndexBucket.Cursor()
	for value, otherID := cursor.First(); value != nil; value, otherID = cursor.Next() {
		if bytes.Equal(otherID, id) {
			return ix.Remove(value)
		}
	}
	return nil
}

// Get the id behind an indexed value.
func (ix *UniqueIndex) Get(indexValue []byte) []byte {
	return ix.IndexBucket.Get(indexValue)
}

// All returns all the IDs corresponding to the given index value.
// For unique indexes this should be a single ID.
func (ix *UniqueIndex) All(indexValue []byte, _ *indexing.CursorOptions) [][]byte {
	id := ix.IndexBucket.Get(indexValue)
	if id != nil {
		return [][]byte{id}
	}
	return nil
}

// AllRecords returns all the IDs.
func (ix *UniqueIndex) AllRecords(ops *indexing.CursorOptions) [][]byte {
	shouldReverse := ops != nil && ops.Reverse
	c := &ReversibleCursor{
		C:       ix.IndexBucket.Cursor(),
		Reverse: shouldReverse,
	}
	return scanCursor(c, ops)
}

// Range gets the IDs in the given range.
func (ix *UniqueIndex) Range(min []byte, max []byte, ops *indexing.CursorOptions) [][]byte {
	shouldReverse := ops != nil && ops.Reverse
	c := &RangeCursor{
		C:          ix.IndexBucket.Cursor(),
		Reverse:    shouldReverse,
		Min:        min,
		Max:        max,
		Comparator: bytes.Compare,
	}
	return scanCursor(c, ops)
}

// AllWithPrefix finds all the IDs that are prefixed with a given byte array
func (ix *UniqueIndex) AllWithPrefix(prefix []byte, ops *indexing.CursorOptions) [][]byte {
	c := &PrefixCursor{
		C:       ix.IndexBucket.Cursor(),
		Reverse: ops != nil && ops.Reverse,
		Prefix:  prefix,
	}
	return scanCursor(c, ops)
}

func (ix *UniqueIndex) IsEmpty() bool {
	c := &ReversibleCursor{
		C:       ix.IndexBucket.Cursor(),
		Reverse: false,
	}
	val, _ := c.First()
	return val != nil
}

func (ix *UniqueIndex) GoOverCursor(action func(id []byte), opts *indexing.CursorOptions) {
	c := &ReversibleCursor{
		C:       ix.IndexBucket.Cursor(),
		Reverse: opts != nil && opts.Reverse,
	}
	for value, id := c.First(); value != nil; value, id = c.Next() {
		if opts != nil && opts.Skip > 0 {
			opts.Skip--
			continue
		}
		if opts != nil && opts.Limit == 0 {
			break
		}
		if opts != nil && opts.Limit > 0 {
			opts.Limit--
		}
		action(id)
	}
}

func scanCursor(c indexing.Cursor, ops *indexing.CursorOptions) [][]byte {
	var results [][]byte
	for value, id := c.First(); value != nil && c.CanContinue(value); value, id = c.Next() {
		if ops != nil && ops.Skip > 0 {
			ops.Skip--
			continue
		}
		if ops != nil && ops.Limit == 0 {
			break
		}
		if ops != nil && ops.Limit > 0 {
			ops.Limit--
		}
		results = append(results, id)
	}
	return results
}
