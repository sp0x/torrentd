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
}

const (
	resultsCollection = "results"
	metaCollection    = "meta"
	metaDocId         = "__meta"
)

//NewFirestoreStorage creates a new firestore backed storage
func NewFirestoreStorage(opts *FirestoreConfig) (*FirestoreStorage, error) {
	ctx := context.Background()
	var options []option.ClientOption
	if opts.CredentialsFile != "" {
		options = append(options, option.WithCredentialsFile(opts.CredentialsFile))
	}
	// credentials file option is optional, by default it will use GOOGLE_APPLICATION_CREDENTIALS
	// environment variable, this is a default method to connect to Google services
	client, err := firestore.NewClient(ctx, opts.ProjectId, options...)
	if err != nil {
		return nil, err
	}
	return &FirestoreStorage{
		context: ctx,
		client:  client,
	}, nil
}

func (f *FirestoreStorage) Find(query indexing.Query, result *search.ExternalResultItem) error {
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

func (f *FirestoreStorage) Create(parts indexing.Key, item *search.ExternalResultItem) error {
	collection := f.client.Collection(metaCollection)
	indexValue := indexing.GetIndexValueFromItem(parts, item)
	_, err := collection.Doc(string(indexValue)).Create(f.context, item)
	if err != nil {
		return err
	}
	//Update the meta, incrementing t he count
	doc, err := collection.Doc(metaDocId).Get(f.context)
	if err != nil && err != iterator.Done {
		return err
	}
	if doc == nil {
		metaDocItem := make(map[string]interface{})
		metaDocItem["count"] = 1
		_, _ = collection.Doc(metaDocId).Create(f.context, metaDocItem)
	} else {
		size := doc.Data()["count"].(int)
		_, err = collection.Doc(metaDocId).Update(f.context, []firestore.Update{
			{Path: "count", Value: size + 1},
		})
	}
	return err
}

//Size is the size of the storage, as in records count
func (f *FirestoreStorage) Size() int64 {
	collection := f.client.Collection(metaCollection)
	doc, err := collection.Doc(metaDocId).Get(f.context)
	if err != nil {
		return -1
	}
	size := doc.Data()["count"]
	return int64(size.(int))
}

//GetNewest returns the latest `count` of records.
func (f *FirestoreStorage) GetNewest(count int) []*search.ExternalResultItem {
	var output []*search.ExternalResultItem
	collection := f.client.Collection(metaCollection)
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
		output = append(output, &newItem)
	}
	return output
}
