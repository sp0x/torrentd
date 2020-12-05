package serializers

import (
	"reflect"
)

type DynamicMarshaler struct {
	recordValue reflect.Value
	recordType  reflect.Type
	marshaler   MarshalUnmarshaler
}

func NewDynamicMarshaler(recordPtr interface{}, underlayingMarshaler MarshalUnmarshaler) *DynamicMarshaler {
	m := &DynamicMarshaler{
		marshaler: underlayingMarshaler,
	}
	m.SetOutputType(recordPtr)
	return m
}

func (d *DynamicMarshaler) SetOutputType(recordPtr interface{}) {
	if reflect.TypeOf(recordPtr).Kind() != reflect.Ptr {
		panic("ptr required")
	}
	d.recordValue = reflect.ValueOf(recordPtr)
	d.recordType = reflect.Indirect(d.recordValue).Type()
}

func (d *DynamicMarshaler) Unmarshal(data []byte) (interface{}, error) {
	record := reflect.New(d.recordType).Interface()
	err := d.marshaler.Unmarshal(data, record)
	return record, err
}

func (d *DynamicMarshaler) UnmarshalAt(data []byte, ptr interface{}) error {
	err := d.marshaler.Unmarshal(data, ptr)
	return err
}

func (d *DynamicMarshaler) Marshal(item interface{}) ([]byte, error) {
	return d.marshaler.Marshal(item)
}

func (d *DynamicMarshaler) New() interface{} {
	record := reflect.New(d.recordType).Interface()
	return record
}
