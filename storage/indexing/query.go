package indexing

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/sp0x/torrentd/indexer/search"
	"reflect"
)

type Key []string

type Query interface {
	Put(k, v interface{})
	Size() int
	Keys() []interface{}
	Values() []interface{}
	Get(key interface{}) (value interface{}, found bool)
}

func NewQuery() Query {
	query := linkedhashmap.New()
	return query
}

//NewKey creates a new keyParts using an array of fields.
func NewKey(fieldNames ...string) Key {
	var key Key
	for _, item := range fieldNames {
		key = append(key, item)
	}
	return key
}

//GetKeyQueryFromItem gets the query that matches an item with the given keyParts.
func GetKeyQueryFromItem(keyParts Key, item *search.ExternalResultItem) Query {
	output := NewQuery()
	val := reflect.ValueOf(item).Elem()
	for _, kfield := range keyParts {
		fld := val.FieldByName(kfield)
		if !fld.IsValid() {
			output.Put(kfield, item.GetField(kfield))
			//output[kfield] = item.GetField(kfield)
		} else {
			output.Put(kfield, fld.Interface())
			//output[kfield] = fld.Interface()
		}
	}
	return output
}
