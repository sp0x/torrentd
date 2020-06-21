package indexing

import (
	"fmt"
	"github.com/sp0x/torrentd/indexer/search"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Index interface {
	Add(value []byte, targetID []byte) error
	Remove(value []byte) error
	RemoveById(id []byte) error
	Get(value []byte) []byte
	All(value []byte, opts *CursorOptions) [][]byte
	AllRecords(opts *CursorOptions) [][]byte
	Range(min []byte, max []byte, opts *CursorOptions) [][]byte
	//	AllWithPrefix(prefix []byte, opts *CursorOptions) ([][]byte)
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
