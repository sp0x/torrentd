package indexing

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"reflect"
	"strings"
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
	key.fieldsCache = make(map[string]interface{})
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
	if k.fieldsCache == nil {
		k.fieldsCache = make(map[string]interface{})
	}
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
func KeyHasValue(key *Key, item interface{}) bool {
	val := reflect.ValueOf(item).Elem()
	fieldsField := val.FieldByName("ExtraFields")
	for _, key := range key.Fields {
		isExtra := strings.HasPrefix(key, "ExtraFields.")
		if isExtra {
			key = key[12:]
		}
		fld := val.FieldByName(key)
		if fld.IsValid() {
			val := fld.Interface()
			if val == nil || val.(string) == "" {
				return false
			}
			continue
		}
		if val, found := fieldsField.Interface().(map[string]interface{})[key]; found {
			if val == nil || val.(string) == "" {
				return false
			}
			continue
		}
		return false
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
func GetKeyQueryFromItem(keyParts *Key, item interface{}) Query {
	output := NewQuery()
	val := reflect.ValueOf(item).Elem()
	fieldsField := val.FieldByName("ExtraFields")
	for _, key := range keyParts.Fields {
		isExtra := strings.HasPrefix(key, "ExtraFields.")
		parsedKey := key
		if isExtra {
			parsedKey = key[12:]
		}
		fld := val.FieldByName(parsedKey)
		if fld.IsValid() {
			val := fld.Interface()
			output.Put(key, val)
			continue
		}
		if val, found := fieldsField.Interface().(map[string]interface{})[parsedKey]; found {
			output.Put(key, val)
		}
	}
	return output
}
