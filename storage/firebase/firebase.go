package firebase

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage/indexing"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirestoreConfig struct {
	ProjectId       string
	CredentialsFile string
}

type FirestoreStorage struct {
	client  *firestore.Client
	context context.Context
	counter counter
}

const (
	resultsCollection = "results"
	//metaCollection    = "meta"
	metaDocId  = "__meta"
	counterDoc = "__count"
)

//NewFirestoreStorage creates a new firestore backed storage
func NewFirestoreStorage(conf *FirestoreConfig) (*FirestoreStorage, error) {
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
	f := &FirestoreStorage{
		context: ctx,
		client:  client,
		counter: counter{20},
	}
	err = f.counter.initCounterIfNeeded(client.Collection(resultsCollection), f.context, counterDoc)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *FirestoreStorage) Find(query indexing.Query, result *search.ExternalResultItem) error {
	//if query.Size()==0 || query==nil{
	//	query = indexing.NewQuery()
	//	query.Put("GUID", result.GUID)
	//}
	fireQuery := f.transformIndexQueryToFirestoreQuery(query, 1)
	documentIterator := fireQuery.Limit(1).Documents(f.context)
	document, err := documentIterator.Next()
	if err != nil {
		return err
	}
	return document.DataTo(result)
}

func (f *FirestoreStorage) transformIndexQueryToFirestoreQuery(query indexing.Query, limit int) *firestore.Query {
	collection := f.client.Collection(resultsCollection)
	var fireQuery *firestore.Query
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

func (f *FirestoreStorage) Update(query indexing.Query, item *search.ExternalResultItem) error {
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
//This uses the GUID for identifying records, upon creation a new UUID is generated.
func (f *FirestoreStorage) Create(item *search.ExternalResultItem) error {
	return f.CreateWithKey(nil, item)
}

//CreateWithKey creates a new record using a custom key.
//If a key isn't provided, a random uuid is generated in it's place, and stored in the GUID field.
func (f *FirestoreStorage) CreateWithKey(key indexing.Key, item *search.ExternalResultItem) error {
	collection := f.client.Collection(resultsCollection)
	indexValue := ""
	var doc *firestore.DocumentRef
	if len(key) == 0 {
		doc = collection.NewDoc()
	} else {
		indexValue = string(indexing.GetIndexValueFromItem(key, item))
		doc = collection.Doc(string(indexValue))
	}
	if len(key) == 0 {
		//Since this is a new item we'll need to create a new ID, if there's no key.
		item.GUID = doc.ID
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
	collection := f.client.Collection(resultsCollection)
	doc, err := collection.Doc(metaDocId).Get(f.context)
	if err != nil {
		return -1
	}
	size := doc.Data()["count"]
	return int64(size.(int))
}

//GetNewest returns the latest `count` of records.
func (f *FirestoreStorage) GetNewest(count int) []search.ExternalResultItem {
	var output []search.ExternalResultItem
	collection := f.client.Collection(resultsCollection)
	iter := collection.OrderBy("ID", firestore.Desc).Limit(count).Documents(f.context)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if doc == nil {
			continue
		}
		newItem := search.ExternalResultItem{}
		err = doc.DataTo(&newItem)
		if err != nil {
			continue
		}
		output = append(output, newItem)
	}
	return output
}
