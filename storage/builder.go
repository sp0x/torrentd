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

type Builder struct {
	backingType string
	primaryKey  *indexing.Key
	endpoint    string
	namespace   string
	backing     ItemStorageBacking
}

func (b *Builder) WithBacking(backingType string) *Builder {
	b.backingType = backingType
	return b
}

func (b *Builder) BackedBy(backing ItemStorageBacking) *Builder {
	b.backing = backing
	return b
}

func (b *Builder) WithDefaultBacking() *Builder {
	b.backingType = "boltdb"
	return b
}

func (b *Builder) WithPK(keyFields *indexing.Key) *Builder {
	b.primaryKey = keyFields
	return b
}

func (b *Builder) WithEndpoint(endpoint string) *Builder {
	b.endpoint = endpoint
	return b
}

func (b *Builder) WithNamespace(ns string) *Builder {
	b.namespace = ns
	return b
}

func (b *Builder) Build() ItemStorage {
	backing := b.backing
	if backing == nil {
		bfn, ok := storageBackingMap[b.backingType]
		if !ok {
			panic("Unsupported storage backing type")
		}
		backing = bfn(b)
	}
	return &KeyedStorage{
		primaryKey:     *b.primaryKey,
		backing:        backing,
		indexKeysCache: make(map[string]interface{}),
	}
}

func init() {
	storageBackingMap["boltdb"] = func(builder *Builder) ItemStorageBacking {
		b, err := bolt.NewBoltStorage(builder.endpoint)
		if err != nil {
			fmt.Printf("Error while constructing boltdb storage: %v", err)
			os.Exit(1)
		}
		if b == nil {
			fmt.Printf("Couldn't construct boltdb storage.")
			os.Exit(1)
		}
		b.SetNamespace(builder.namespace)
		return b
	}
	storageBackingMap["firebase"] = func(builder *Builder) ItemStorageBacking {
		conf := &firebase.FirestoreConfig{Namespace: builder.namespace}
		conf.ProjectId = viper.Get("firebase_project").(string)
		conf.CredentialsFile = viper.Get("firebase_credentials_file").(string)
		b, err := firebase.NewFirestoreStorage(conf)
		if err != nil {
			log.Error(err)
			return nil
		}
		return b
	}
	storageBackingMap["sqlite"] = func(builder *Builder) ItemStorageBacking {
		panic("Deprecated")
		//b := &sqlite.DBStorage{}
		//return b
	}
}
