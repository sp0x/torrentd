package bolt

import (
	"fmt"
	"github.com/boltdb/bolt"
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/storage/serializers/gob"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUniqueIndex(t *testing.T) {
	g := NewWithT(t)
	db, _ := GetBoltDb(tempfile())
	defer func() {
		_ = db.Close()
	}()

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := NewUniqueIndex(b, []byte("uindex1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id2"))
		require.Error(t, err)
		g.Expect(err).ToNot(BeNil())

		err = idx.Add(nil, []byte("id2"))
		require.Error(t, err)
		g.Expect(err).ToNot(BeNil())

		err = idx.Add([]byte("hi"), nil)
		require.Error(t, err)
		g.Expect(err).ToNot(BeNil())

		id := idx.Get([]byte("hello"))
		require.Equal(t, []byte("id1"), id)

		id = idx.Get([]byte("goodbye"))
		require.Nil(t, id)

		err = idx.Remove([]byte("hello"))
		require.NoError(t, err)

		err = idx.Remove(nil)
		require.NoError(t, err)

		id = idx.Get([]byte("hello"))
		require.Nil(t, id)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hi"), []byte("id2"))
		require.NoError(t, err)

		err = idx.Add([]byte("yo"), []byte("id3"))
		require.NoError(t, err)

		list := idx.AllRecords(nil)
		require.NoError(t, err)
		require.Len(t, list, 3)

		opts := indexing.NewCursorOptions()
		opts.Limit = 2
		list = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, list, 2)

		opts = indexing.NewCursorOptions()
		opts.Skip = 2
		list = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, []byte("id3"), list[0])

		opts = indexing.NewCursorOptions()
		opts.Skip = 2
		opts.Limit = 1
		opts.Reverse = true
		list = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, []byte("id1"), list[0])

		err = idx.RemoveById([]byte("id2"))
		require.NoError(t, err)

		id = idx.Get([]byte("hello"))
		require.Equal(t, []byte("id1"), id)
		id = idx.Get([]byte("hi"))
		require.Nil(t, id)
		id = idx.Get([]byte("yo"))
		require.Equal(t, []byte("id3"), id)
		ids := idx.All([]byte("yo"), nil)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id3"), ids[0])

		err = idx.RemoveById([]byte("id2"))
		require.NoError(t, err)
		err = idx.RemoveById([]byte("id4"))
		require.NoError(t, err)
		return nil
	})

	require.NoError(t, err)
}

func TestUniqueIndexRange(t *testing.T) {
	db, _ := GetBoltDb(tempfile())
	defer func() {
		_ = db.Close()
	}()
	gobs := gob.Serializer

	_ = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := NewUniqueIndex(b, []byte("uindex1"))
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			val, _ := gobs.Marshal(i)
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		min, _ := gobs.Marshal(3)
		max, _ := gobs.Marshal(5)
		list := idx.Range(min, max, nil)
		require.Len(t, list, 3)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4, 5}, list)

		min, _ = gobs.Marshal(11)
		max, _ = gobs.Marshal(20)
		list = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gobs.Marshal(7)
		max, _ = gobs.Marshal(2)
		list = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gobs.Marshal(-5)
		max, _ = gobs.Marshal(2)
		list = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gobs.Marshal(3)
		max, _ = gobs.Marshal(7)
		opts := indexing.NewCursorOptions()
		opts.Skip = 2
		list = idx.Range(min, max, opts)
		require.Len(t, list, 3)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 6, 7}, list)

		opts = indexing.NewCursorOptions()
		opts.Limit = 2
		list = idx.Range(min, max, opts)
		require.Len(t, list, 2)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4}, list)

		opts = indexing.NewCursorOptions()
		opts.Reverse = true
		opts.Skip = 2
		opts.Limit = 2
		list = idx.Range(min, max, opts)
		require.Len(t, list, 2)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 4}, list)
		return nil
	})
}

func TestUniqueIndexPrefix(t *testing.T) {
	db, _ := GetBoltDb(tempfile())
	defer func() {
		_ = db.Close()
	}()

	_ = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := NewUniqueIndex(b, []byte("uindex1"))
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			val := []byte(fmt.Sprintf("a%d", i))
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		for i := 0; i < 10; i++ {
			val := []byte(fmt.Sprintf("b%d", i))
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		list := idx.AllWithPrefix([]byte("a"), nil)
		require.Len(t, list, 10)
		require.NoError(t, err)

		list = idx.AllWithPrefix([]byte("b"), nil)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("b0"), list[0])
		require.Equal(t, []byte("b9"), list[9])

		opts := indexing.NewCursorOptions()
		opts.Reverse = true
		list = idx.AllWithPrefix([]byte("a"), opts)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("a9"), list[0])
		require.Equal(t, []byte("a0"), list[9])

		opts = indexing.NewCursorOptions()
		opts.Reverse = true
		list = idx.AllWithPrefix([]byte("b"), opts)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("b9"), list[0])
		require.Equal(t, []byte("b0"), list[9])

		opts = indexing.NewCursorOptions()
		opts.Skip = 9
		opts.Limit = 5
		list = idx.AllWithPrefix([]byte("a"), opts)
		require.Len(t, list, 1)
		require.NoError(t, err)
		require.Equal(t, []byte("a9"), list[0])

		opts = indexing.NewCursorOptions()
		opts.Reverse = true
		opts.Skip = 9
		opts.Limit = 5
		list = idx.AllWithPrefix([]byte("a"), opts)
		require.Len(t, list, 1)
		require.NoError(t, err)
		require.Equal(t, []byte("a0"), list[0])
		return nil
	})
}

func assertEncodedIntListEqual(t *testing.T, expected []int, actual [][]byte) {
	ints := make([]int, len(actual))

	for i, e := range actual {
		err := gob.Serializer.Unmarshal(e, &ints[i])
		require.NoError(t, err)
	}

	require.Equal(t, expected, ints)
}
