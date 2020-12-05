package firebase

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/storage/serializers"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirestoreConfig struct {
	ProjectId       string
	CredentialsFile string
	Namespace       string
}

type FirestoreStorage struct {
	client    *firestore.Client
	context   context.Context
	counter   counter
	namespace string
	marshaler *serializers.DynamicMarshaler
}

const (
	resultsCollection = "results"
	metaDoc           = "meta"
	counterDoc        = "__count"
)

//NewFirestoreStorage creates a new firestore backed storage
func NewFirestoreStorage(conf *FirestoreConfig, typePtr interface{}) (*FirestoreStorage, error) {
	ctx := context.Background()
	var options []option.ClientOption
	if conf.CredentialsFile != "" {
		options = append(options, option.WithCredentialsFile(conf.CredentialsFile))
	}
	// credentials file option is optional, by default it will use GOOGLE_APPLICATION_CREDENTIALS
	// environment variable, this is a default method to connect to Google services
	client, err := firestore.NewClient(ctx, conf.ProjectId, options...)
	if err != nil {
		return nil, err
	}
	targetCollection := conf.Namespace
	if targetCollection == "" {
		targetCollection = resultsCollection
	}
	docCounter := counter{20}
	f := &FirestoreStorage{
		context:   ctx,
		client:    client,
		counter:   docCounter,
		namespace: targetCollection,
		marshaler: serializers.NewDynamicMarshaler(typePtr, nil),
	}
	err = f.counter.initCounterIfNeeded(f.getCollection(), f.context, counterDoc)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *FirestoreStorage) getCollection() *firestore.CollectionRef {
	return f.client.Collection(f.namespace)
}

func (f *FirestoreStorage) Close() {
	//This is just a stub
}

func (f *FirestoreStorage) HasIndex(meta *indexing.IndexMetadata) bool {
	return false
}

func (f *FirestoreStorage) GetIndexes() map[string]indexing.IndexMetadata {
	return nil
}

func (f *FirestoreStorage) Find(query indexing.Query, result interface{}) error {
	fireQuery := f.transformIndexQueryToFirestoreQuery(query, 1)
	documentIterator := fireQuery.Limit(1).Documents(f.context)
	document, err := documentIterator.Next()
	if err != nil {
		return err
	}
	return document.DataTo(result)
}

func (f *FirestoreStorage) ForEach(callback func(record search.ResultItemBase)) {
	fireQuery := f.transformIndexQueryToFirestoreQuery(nil, 1)
	documentIterator := fireQuery.Documents(f.context)
	for true {
		document, err := documentIterator.Next()
		if err == iterator.Done {
			break
		}
		if document == nil {
			continue
		}
		record := f.marshaler.New()
		err = document.DataTo(record)
		if err != nil {
			break
		}
		callback(record.(search.ResultItemBase))
	}

}

func (f *FirestoreStorage) transformIndexQueryToFirestoreQuery(query indexing.Query, limit int) *firestore.Query {
	collection := f.getCollection()
	var fireQuery *firestore.Query
	if query == nil {
		tmpQuery := collection.Offset(0).Limit(limit)
		return &tmpQuery
	}
	for _, key := range query.Keys() {
		val, _ := query.Get(key)
		if fireQuery != nil {
			crQuery := fireQuery.Where(key.(string), "==", val)
			fireQuery = &crQuery
		} else {
			crQuery := collection.Where(key.(string), "==", val)
			fireQuery = &crQuery
		}
	}
	if fireQuery == nil {
		crQuery := collection.Limit(limit)
		fireQuery = &crQuery
	}
	return fireQuery
}

func (f *FirestoreStorage) Update(query indexing.Query, item interface{}) error {
	fireQuery := f.transformIndexQueryToFirestoreQuery(query, 1)
	docs := fireQuery.Documents(f.context)
	firstDoc, err := docs.Next()
	if err != nil {
		return err
	}
	_, err = firstDoc.Ref.Set(f.context, item)
	return err
}

//Create a new record.
//This uses the UUIDValue for identifying records, upon creation a new UUID is generated.
func (f *FirestoreStorage) Create(item search.Record, additionalIndex *indexing.Key) error {
	err := f.CreateWithId(nil, item, additionalIndex)
	if err != nil {
		return err
	}
	if additionalIndex.IsEmpty() {
		return nil
	}
	//Right now we don't do anything with that index.....
	return err
}

//CreateWithId creates a new record using a custom key.
//If a key isn't provided, a random uuid is generated in it's place, and stored in the UUIDValue field.
func (f *FirestoreStorage) CreateWithId(key *indexing.Key, item search.Record, uniqueIndexKeys *indexing.Key) error {
	collection := f.getCollection()
	indexValue := ""
	var doc *firestore.DocumentRef
	if key == nil || key.IsEmpty() {
		doc = collection.NewDoc()
	} else {
		indexValue = string(indexing.GetIndexValueFromItem(key, item))
		if indexValue == "" {
			return errors.New("id was empty, it's required for firebase")
		}
		doc = collection.Doc(indexValue)
	}
	if key == nil || key.IsEmpty() {
		//Since this is a new item we'll need to create a new ID, if there's no key.
		item.SetUUID(doc.ID)
	}
	_, err := doc.Create(f.context, item)
	if err != nil {
		return err
	}
	//Update the meta, incrementing the count
	_, err = f.counter.incrementCounter(f.context, collection.Doc(counterDoc))
	return err
}

//Size is the size of the storage, as in records count
func (f *FirestoreStorage) Size() int64 {
	collection := f.getCollection()
	doc, err := collection.Doc(metaDoc).Get(f.context)
	if err != nil {
		return -1
	}
	size := doc.Data()["count"]
	return int64(size.(int))
}
