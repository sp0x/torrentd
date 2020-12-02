package storage

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/storage/bolt"
	"github.com/sp0x/torrentd/storage/firebase"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/spf13/viper"
	"os"
)

var storageBackingMap = make(map[string]func(builder *Builder) ItemStorageBacking)

func NewBuilder() *Builder {
	b := &Builder{}
	return b.WithDefaultBacking()
}

//Builder for ItemStorage
type Builder struct {
	backingType   string
	primaryKey    *indexing.Key
	endpoint      string
	namespace     string
	backing       ItemStorageBacking
	recordTypePtr interface{}
}

//WithBacking set the data backing type. ex: boltdb, firebase, sqlite.
func (b *Builder) WithBacking(backingType string) *Builder {
	b.backingType = backingType
	return b
}

//WithRecord set the type of the items you'll be working with, this is needed for unmarshaling and querying.
func (b *Builder) WithRecord(recordTypePtr interface{}) *Builder {
	b.recordTypePtr = recordTypePtr
	return b
}

//BackedBy can be used if you already have an initialized storage backing.
func (b *Builder) BackedBy(backing ItemStorageBacking) *Builder {
	b.backing = backing
	return b
}

//WithDefaultBacking sets the storage backing to `boltdb`
func (b *Builder) WithDefaultBacking() *Builder {
	b.backingType = "boltdb"
	return b
}

//WithPK use a primary key to store the data. By default an UUID is used.
func (b *Builder) WithPK(keyFields *indexing.Key) *Builder {
	b.primaryKey = keyFields
	return b
}

//WithEndpoint if your storage location has a specific location (ex: db path)
func (b *Builder) WithEndpoint(endpoint string) *Builder {
	b.endpoint = endpoint
	return b
}

//WithNamespace make sure all data is in that namespace
func (b *Builder) WithNamespace(ns string) *Builder {
	if ns == "" {
		panic("namespace must be non-empty")
	}
	b.namespace = ns
	return b
}

//Build the storage object
func (b *Builder) Build() ItemStorage {
	backing := b.backing
	if b.recordTypePtr == nil {
		panic("record type is required")
	}
	if backing == nil {
		storageResolverFunc, ok := storageBackingMap[b.backingType]
		if !ok {
			var supportedStorages []string
			for k := range storageBackingMap {
				supportedStorages = append(supportedStorages, k)
			}
			log.WithFields(log.Fields{"requested": b.backingType, "supported": supportedStorages}).
				Error("Unsupported storage backing type.")
			panic("Unsupported storage backing type.")
		}
		backing = storageResolverFunc(b)
	}
	key := b.primaryKey
	if key == nil {
		key = &indexing.Key{}
	}
	return &KeyedStorage{
		primaryKey:     *key,
		backing:        backing,
		indexKeysCache: make(map[string]interface{}),
	}
}

func init() {
	storageBackingMap["boltdb"] = func(builder *Builder) ItemStorageBacking {
		b, err := bolt.NewBoltStorage(builder.endpoint, builder.recordTypePtr)
		if err != nil {
			fmt.Printf("Error while constructing boltdb storage: %v", err)
			os.Exit(1)
		}
		if b == nil {
			fmt.Printf("Couldn't construct boltdb storage.")
			os.Exit(1)
		}
		//Set to the namespace of the builder(this may be the index)
		err = b.SetNamespace(builder.namespace)
		if err != nil {
			os.Exit(1)
		}

		return b
	}
	storageBackingMap["firebase"] = func(builder *Builder) ItemStorageBacking {
		conf := &firebase.FirestoreConfig{Namespace: builder.namespace}
		conf.ProjectId = viper.Get("firebase_project").(string)
		conf.CredentialsFile = viper.Get("firebase_credentials_file").(string)
		b, err := firebase.NewFirestoreStorage(conf, builder.recordTypePtr)
		if err != nil {
			log.Error(err)
			return nil
		}
		return b
	}
	storageBackingMap["sqlite"] = func(builder *Builder) ItemStorageBacking {
		panic("sqlite storage is deprecated and shouldn't be used anymore")
		//b := &sqlite.DBStorage{}
		//return b
	}
}
