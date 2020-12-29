package bolt

import (
	"bytes"

	"github.com/boltdb/bolt"
)

type FilteredCursor struct {
	C       *bolt.Cursor
	Filters []func([]byte, []byte) bool
}

func shouldSkip(fc *FilteredCursor, id []byte, value []byte) bool {
	for _, filter := range fc.Filters {
		isFiltered := filter(id, value)
		if isFiltered {
			return true
		}
	}
	return false
}

func (c *FilteredCursor) First() ([]byte, []byte) {
	id, value := c.C.First()
	for true {
		if !shouldSkip(c, id, value) {
			break
		}
		id, value = c.Next()
	}
	return id, value
}

func (c *FilteredCursor) Next() ([]byte, []byte) {
	id, value := c.C.Next()
	for true {
		if !shouldSkip(c, id, value) {
			break
		}
		id, value = c.Next()
	}
	return id, value
}

func (c *FilteredCursor) CanContinue(val []byte) bool {
	return val != nil
}

type ReversibleCursor struct {
	C       *bolt.Cursor
	Reverse bool
}

func (c *ReversibleCursor) First() ([]byte, []byte) {
	if c.Reverse {
		return c.C.Last()
	}
	return c.C.First()
}

func (c *ReversibleCursor) Next() ([]byte, []byte) {
	if c.Reverse {
		return c.C.Prev()
	}
	return c.C.Next()
}

func (c *ReversibleCursor) CanContinue(val []byte) bool {
	return val != nil
}

type RangeCursor struct {
	C          *bolt.Cursor
	Reverse    bool
	Min        []byte
	Max        []byte
	Comparator func(a []byte, b []byte) int
}

// First gets the first element in the range.
func (r *RangeCursor) First() ([]byte, []byte) {
	if !r.Reverse {
		return r.C.Seek(r.Min)
	}
	// Seek to the end
	key, val := r.C.Seek(r.Max)
	// Go one step back
	if !bytes.HasPrefix(key, r.Max) && key != nil {
		key, val = r.C.Prev()
	}
	return key, val
}

// CanContinue checks if the cursor can continue to the given value
func (r *RangeCursor) CanContinue(value []byte) bool {
	if r.Reverse {
		return value != nil && r.Comparator(value, r.Min) >= 0
	}
	return value != nil && r.Comparator(value, r.Max) <= 0
}

// Next gets the next element in the cursor.
func (r *RangeCursor) Next() ([]byte, []byte) {
	if r.Reverse {
		return r.C.Prev()
	}
	return r.C.Next()
}

type PrefixCursor struct {
	C       *bolt.Cursor
	Reverse bool
	Prefix  []byte
}

// First item in the cursor that matches the prefix.
func (c *PrefixCursor) First() ([]byte, []byte) {
	var key, val []byte
	key, val = c.C.Seek(c.Prefix)
	if key == nil {
		return nil, nil
	}
	if !c.Reverse {
		return key, val
	}
	kc, vc := key, val
	for ; kc != nil && bytes.HasPrefix(kc, c.Prefix); kc, vc = c.C.Next() {
		key, val = kc, vc
	}
	if kc != nil {
		key, val = c.C.Prev()
	}
	return key, val
}

// CanContinue figures out if the cursor can continue
func (c *PrefixCursor) CanContinue(value []byte) bool {
	return value != nil && bytes.HasPrefix(value, c.Prefix)
}

// Next item in the cursor
func (c *PrefixCursor) Next() ([]byte, []byte) {
	if c.Reverse {
		return c.C.Prev()
	}
	return c.C.Next()
}
