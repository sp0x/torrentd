package indexing

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/sp0x/torrentd/indexer/search"
	"reflect"
)

//Key is a primary key or an indexing key, this can be a composite key as well
//[]string
type Key struct {
	Fields      []string
	fieldsCache map[string]interface{}
	//This can be used to prefix the value of the key.
	//ValuePrefix string
}

func (k *Key) IsEmpty() bool {
	return len(k.Fields) == 0
}

//NewKey creates a new keyParts using an array of fields.
func NewKey(fieldNames ...string) *Key {
	var key Key
	for _, item := range fieldNames {
		_, exists := key.fieldsCache[item]
		if exists {
			continue
		}
		key.Fields = append(key.Fields, item)
		key.fieldsCache[item] = true
	}
	return &key
}

//func (k *Key) SetPrefix(val string) *Key {
//	k.ValuePrefix = val
//	return k
//}

//AddKeys adds multiple keys
func (k *Key) AddKeys(newKeys *Key) {
	for _, newKey := range newKeys.Fields {
		_, exists := k.fieldsCache[newKey]
		if exists {
			continue
		}
		k.Fields = append(k.Fields, newKey)
		k.fieldsCache[newKey] = true
	}
}

//Add a new key field
func (k *Key) Add(s string) {
	_, exists := k.fieldsCache[s]
	if exists {
		return
	}
	k.Fields = append(k.Fields, s)
	k.fieldsCache[s] = true
}

//KeyHasValue checks if all the key fields in an item have a value.
func KeyHasValue(key *Key, item *search.ExternalResultItem) bool {
	val := reflect.ValueOf(item).Elem()
	for _, key := range key.Fields {
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

//GetKeyQueryFromItem gets the query that matches an item with the given keyParts.
func GetKeyQueryFromItem(keyParts *Key, item *search.ExternalResultItem) Query {
	output := NewQuery()
	val := reflect.ValueOf(item).Elem()
	for _, kfield := range keyParts.Fields {
		fld := val.FieldByName(kfield)
		if !fld.IsValid() {
			output.Put(kfield, item.GetField(kfield))
		} else {
			output.Put(kfield, fld.Interface())
		}
	}
	return output
}
