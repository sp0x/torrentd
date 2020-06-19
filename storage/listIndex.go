package storage

import (
	"bytes"
	"errors"
	"github.com/boltdb/bolt"
)

const (
	idsIndex = "__ids"
)

//NewListIndex creates a new list index with it's sub-buckets.
func NewListIndex(parent *bolt.Bucket, name []byte) (*ListIndex, error) {
	var err error
	b := parent.Bucket(name)
	if b != nil {
		if !parent.Writable() {
			return nil, errors.New("parent bucket is not writable")
		}
		b, err = parent.CreateBucket(name)
		if err != nil {
			return nil, err
		}
	}
	idIndex, err := NewUniqueIndex(b, []byte(idsIndex))
	if err != nil {
		return nil, err
	}
	return &ListIndex{
		IndexBucket:  b,
		ParentBucket: parent,
		IDs:          idIndex,
	}, nil
}

//ListIndex is an index that references values and the matching IDs
type ListIndex struct {
	ParentBucket *bolt.Bucket
	//Bucket that contains all the index values, with values being the ID
	IndexBucket *bolt.Bucket
	//An index that contains all of our IDs, with values being the index values
	IDs *UniqueIndex
}

//Add a new index for an id.
func (ix *ListIndex) Add(val []byte, id []byte) error {
	if val == nil || id == nil {
		return errors.New("value and id are required")
	}
	//Figure out if the id already exists, if it does we'll remove it.
	idKey := ix.IDs.Get(id)
	if idKey != nil {
		err := ix.IndexBucket.Delete(idKey)
		if err != nil {
			return err
		}
		err = ix.IDs.Remove(id)
		if err != nil {
			return err
		}
		//Clear out the id
		idKey = idKey[:0]
	}
	//We create the new key
	idKey = generatePrefix(val)
	idKey = append(idKey, id...)
	err := ix.IDs.Add(id, idKey)
	if err != nil {
		return err
	}
	return ix.IndexBucket.Put(idKey, id)
}

//List indexes are formed like this: <index value>__<id>
func generatePrefix(indexValue []byte) []byte {
	prefix := make([]byte, len(indexValue)+2)
	var i int
	for i = range indexValue {
		prefix[i] = indexValue[i]
	}
	prefix[i+1] = '_'
	prefix[i+2] = '_'
	return prefix
}

//Remove an index
func (ix *ListIndex) Remove(indexValue []byte) error {
	var err error
	var indexKeysToBeDeleted [][]byte
	c := ix.IndexBucket.Cursor()
	prefix := generatePrefix(indexValue)
	for key, _ := c.Seek(prefix); bytes.HasPrefix(key, prefix); key, _ = c.Next() {
		indexKeysToBeDeleted = append(indexKeysToBeDeleted, key)
	}
	//Remove all the indexes.
	for _, key := range indexKeysToBeDeleted {
		err = ix.IndexBucket.Delete(key)
		if err != nil {
			return err
		}
	}
	return ix.IDs.RemoveById(indexValue)
}

//RemoveById removes an index and the matching ID using an ID.
func (ix *ListIndex) RemoveById(id []byte) error {
	//We get the index value
	indexValue := ix.IDs.Get(id)
	//We don't have that index
	if indexValue == nil {
		return nil
	}
	//We delete the index
	err := ix.IndexBucket.Delete(indexValue)
	if err != nil {
		return err
	}
	return ix.IDs.Remove(id)
}

//Get the first ID corresponding to the given value
func (ix *ListIndex) Get(indexValue []byte) []byte {
	c := ix.IndexBucket.Cursor()
	prefix := generatePrefix(indexValue)
	for key, id := c.Seek(prefix); bytes.HasPrefix(key, prefix); {
		return id
	}
	return nil
}

// All the IDs corresponding to the given value
func (ix *ListIndex) All(indexValue []byte, opts *CursorOptions) [][]byte {
	var results [][]byte
	indexCursor := ix.IndexBucket.Cursor()
	cur := ReversibleCursor{
		C:       indexCursor,
		Reverse: opts != nil && opts.Reverse,
	}
	//All IDs are prefixed with the index value
	prefix := generatePrefix(indexValue)
	k, id := indexCursor.Seek(prefix)
	if cur.Reverse {
		var count int
		kc := k
		idc := id
		for ; kc != nil && bytes.HasPrefix(kc, prefix); kc, idc = indexCursor.Next() {
			count++
			k, id = kc, idc
		}
		if kc != nil {
			k, id = indexCursor.Prev()
		}
		results = make([][]byte, 0, count)
	}
	//While the cursor is in our prefixed area.
	for ; bytes.HasPrefix(k, prefix); k, id = cur.Next() {
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
		results = append(results, id)
	}

	return results
}

// AllRecords returns all the IDs of this index
func (ix *ListIndex) AllRecords(opts *CursorOptions) [][]byte {
	var list [][]byte

	c := ReversibleCursor{C: ix.IndexBucket.Cursor(), Reverse: opts != nil && opts.Reverse}

	for k, id := c.First(); k != nil; k, id = c.Next() {
		if id == nil || bytes.Equal(k, []byte("__ids")) {
			continue
		}
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
		list = append(list, id)
	}

	return list
}

// Range returns the ids corresponding to the given range of values
func (ix *ListIndex) Range(min []byte, max []byte, opts *CursorOptions) [][]byte {
	var list [][]byte

	c := RangeCursor{
		C:       ix.IndexBucket.Cursor(),
		Reverse: opts != nil && opts.Reverse,
		Min:     min,
		Max:     max,
		Comparator: func(val, limit []byte) int {
			pos := bytes.LastIndex(val, []byte("__"))
			return bytes.Compare(val[:pos], limit)
		},
	}

	for k, id := c.First(); c.CanContinue(k); k, id = c.Next() {
		if id == nil || bytes.Equal(k, []byte("__ids")) {
			continue
		}
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
		list = append(list, id)
	}

	return list
}

// Prefix returns the ids whose values have the given prefix.
func (ix *ListIndex) Prefix(prefix []byte, opts *CursorOptions) [][]byte {
	var list [][]byte

	c := PrefixCursor{
		C:       ix.IndexBucket.Cursor(),
		Reverse: opts != nil && opts.Reverse,
		Prefix:  prefix,
	}

	for k, id := c.First(); k != nil && c.CanContinue(k); k, id = c.Next() {
		if id == nil || bytes.Equal(k, []byte("__ids")) {
			continue
		}
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
		list = append(list, id)
	}
	return list
}
