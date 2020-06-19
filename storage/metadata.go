package storage

//
//
//
//import (
//	"errors"
//	"fmt"
//	"github.com/boltdb/bolt"
//	"github.com/sp0x/torrentd/storage/serializers"
//	"reflect"
//)
//
//const (
//	metaCodecField = "codec"
//	metadataBucket = "__metadata"
//)
//
//func newMeta(b *bolt.Bucket, marshaler serializers.MarshalUnmarshaler) (*meta, error) {
//	if marshaler == nil{
//		return nil, errors.New("marshaler is required in order to access metadata")
//	}
//	m := b.Bucket([]byte(metadataBucket))
//	if m != nil {
//		name := m.Get([]byte(metaCodecField))
//		if string(name) != marshaler.Name() {
//			return nil, fmt.Errorf("metadata is using `%s` not `%s`, for serialization", name, marshaler.Name())
//		}
//		return &meta{
//			//node:   n,
//			bucket: m,
//		}, nil
//	}
//
//	m, err := b.CreateBucket([]byte(metadataBucket))
//	if err != nil {
//		return nil, err
//	}
//
//	err = m.Put([]byte(metaCodecField), []byte(marshaler.Name()))
//	if err != nil{
//		return nil, err
//	}
//	return &meta{
//		//node:   n,
//		bucket: m,
//	}, nil
//}
//
//type meta struct {
//	//node   Node
//	bucket *bolt.Bucket
//}
//
//func (m *meta) increment(field *fieldConfig) error {
//	var err error
//	counter := field.IncrementStart
//
//	raw := m.bucket.Get([]byte(field.Name + "counter"))
//	if raw != nil {
//		counter, err = numberfromb(raw)
//		if err != nil {
//			return err
//		}
//		counter++
//	}
//
//	raw, err = numbertob(counter)
//	if err != nil {
//		return err
//	}
//
//	err = m.bucket.Put([]byte(field.Name+"counter"), raw)
//	if err != nil {
//		return err
//	}
//
//	field.Value.Set(reflect.ValueOf(counter).Convert(field.Value.Type()))
//	field.IsZero = false
//	return nil
//}
//
