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

//KeyHasValue checks if all the key fields in an item have a value.
func KeyHasValue(key Key, item *search.ExternalResultItem) bool {
	val := reflect.ValueOf(item).Elem()
	for _, key := range key {
		fld := val.FieldByName(key)
		var val interface{}
		if !fld.IsValid() {
			val = item.GetField(key)
		} else {
			val = fld.Interface()
		}
		if val == nil || val.(string) == "" {
			return false
		}
	}
	return true
}

//GetKeyQueryFromItem gets the query that matches an item with the given keyParts.
func GetKeyQueryFromItem(keyParts Key, item *search.ExternalResultItem) Query {
	output := NewQuery()
	val := reflect.ValueOf(item).Elem()
	for _, kfield := range keyParts {
		fld := val.FieldByName(kfield)
		if !fld.IsValid() {
			output.Put(kfield, item.GetField(kfield))
		} else {
			output.Put(kfield, fld.Interface())
		}
	}
	return output
}
