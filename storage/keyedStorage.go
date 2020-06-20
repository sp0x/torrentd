package storage

import (
	"fmt"
	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/prometheus/common/log"
	"github.com/sp0x/torrentd/indexer/search"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type KeyedStorage struct {
	//backing *DBStorage
	backing  ItemStorageBacking
	keyParts Key
}

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

//NewKeyedStorage creates a new keyed storage with the default storage backing.
func NewKeyedStorage(keyFields Key) *KeyedStorage {
	return &KeyedStorage{
		keyParts: keyFields,
		backing:  DefaultStorageBacking(),
	}
}

//NewKeyedStorageWithBacking creates a new keyed storage with a custom storage backing.
func NewKeyedStorageWithBacking(key Key, storage ItemStorageBacking) *KeyedStorage {
	return &KeyedStorage{
		keyParts: key,
		backing:  storage,
	}
}

//NewWithKey gets a storage backed in the same way, with a different key.
func (s *KeyedStorage) NewWithKey(key Key) ItemStorage {
	storage := s.backing

	return &KeyedStorage{
		keyParts: key,
		backing:  storage,
	}
}

//Add handles the discovery of the result, adding additional information like staleness state.
func (s *KeyedStorage) Add(item *search.ExternalResultItem) (bool, bool) {
	var existingResult *search.ExternalResultItem
	existingKey := GetKeyQueryFromItem(s.keyParts, item)
	if existingKey != nil {
		tmpResult := search.ExternalResultItem{}
		if s.backing.Find(existingKey, &tmpResult) == nil {
			existingResult = &tmpResult
		}
	}
	isNew := false
	isUpdate := false
	if existingResult == nil {
		isNew = true
		item.Fingerprint = search.GetResultFingerprint(item)
		err := s.backing.Create(s.keyParts, item)
		if err != nil {
			log.Error(err)
			return false, false
		}
	} else if !existingResult.Equals(item) {
		//This must be an update
		isUpdate = true
		item.Fingerprint = existingResult.Fingerprint
		err := s.backing.Update(existingKey, item)
		if err != nil {
			log.Error(err)
			return false, false
		}
	}
	//We set the result's state so it's known later on whenever it's used.
	item.SetState(isNew, isUpdate)
	return isNew, isUpdate
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

//GetIndexNameFromQuery gets the name of an index from a query.
func GetIndexNameFromQuery(query Query) string {
	name := ""
	querySize := query.Size()
	ix := 0
	for _, key := range query.Keys() {
		name += key.(string)
		if ix < (querySize - 1) {
			name += "_"
		}
	}
	return name
}

//GetIndexValueFromItem gets the index value from a key set and an item.
func GetIndexValueFromItem(keyParts Key, item *search.ExternalResultItem) []byte {
	val := reflect.ValueOf(item).Elem()
	valueParts := make([]string, len(keyParts))
	for ix, kfield := range keyParts {
		fld := val.FieldByName(kfield)
		if !fld.IsValid() {
			valueParts[ix] = serializeKeyValue(item.GetField(kfield))
		} else {
			valueParts[ix] = serializeKeyValue(fld.Interface())
		}
	}
	output := strings.Join(valueParts, "\000")
	return []byte(output)
}

//GetIndexValueFromQuery get the value of an index by a query.
func GetIndexValueFromQuery(query Query) []byte {
	//indexValue := make([]byte, 0, 0)
	valueParts := make([]string, query.Size())
	i := 0
	for _, v := range query.Values() {
		valueParts[i] = serializeKeyValue(v)
		i++
	}
	output := strings.Join(valueParts, "\000")
	return []byte(output)
}

func serializeKeyValue(val interface{}) string {
	switch castVal := val.(type) {
	case string:
		return castVal
	case int:
		return fmt.Sprintf("%v", castVal)
	case int64:
		return fmt.Sprintf("%v", castVal)
	case int16:
		return fmt.Sprintf("%v", castVal)
	case uint:
		return fmt.Sprintf("%v", castVal)
	case uint16:
		return fmt.Sprintf("%v", castVal)
	case uint64:
		return fmt.Sprintf("%v", castVal)
	case float32:
		return fmt.Sprintf("%v", castVal)
	case float64:
		return fmt.Sprintf("%v", castVal)
	case rune:
		return string(castVal)
	case time.Time:
		return fmt.Sprintf("%v", castVal.Unix())
	case bool:
		return strconv.FormatBool(castVal)
	default:
		panic("non supported index value part")
	}
}
