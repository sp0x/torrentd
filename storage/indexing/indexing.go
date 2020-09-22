package indexing

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Index interface {
	Add(value []byte, targetID []byte) error
	Remove(value []byte) error
	RemoveById(id []byte) error
	Get(indexValue []byte) []byte
	// All returns all the IDs corresponding to the given index value.
	All(indexValue []byte, opts *CursorOptions) [][]byte
	AllRecords(opts *CursorOptions) [][]byte
	Range(min []byte, max []byte, opts *CursorOptions) [][]byte
	GoOverCursor(action func(id []byte), opts *CursorOptions)
	//	AllWithPrefix(prefix []byte, opts *CursorOptions) ([][]byte)
}

//IndexMetadata is used to describe an index
type IndexMetadata struct {
	Name     string `json:"name"`
	Unique   bool   `json:"unique"`
	Location string
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
func GetIndexValueFromItem(keyParts *Key, item interface{}) []byte {
	if keyParts == nil {
		return []byte{}
	}
	val := reflect.ValueOf(item)
	element := val.Elem()
	valueParts := make([]string, len(keyParts.Fields))
	fieldsField := element.FieldByName("ExtraFields")
	for ix, fieldName := range keyParts.Fields {
		parsedFieldName := fieldName
		isExtra := strings.HasPrefix(fieldName, "ExtraFields.")
		if isExtra {
			parsedFieldName = parsedFieldName[12:]
		}
		fld := element.FieldByName(fieldName)
		if fld.IsValid() {
			valueParts[ix] = serializeKeyValue(fld.Interface())
			continue
		}
		method := val.MethodByName(parsedFieldName)
		if method.IsValid() {
			rawval := method.Call([]reflect.Value{})[0].Interface()
			valueParts[ix] = serializeKeyValue(rawval)
			continue
		}
		if !fieldsField.IsValid() {
			continue
		}
		if value, found := fieldsField.Interface().(map[string]interface{})[parsedFieldName]; found {
			valueParts[ix] = serializeKeyValue(value)
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
		panic(fmt.Sprintf("non supported index value part: %s", val))
	}
}
