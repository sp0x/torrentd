package firebase

import (
	"cloud.google.com/go/firestore"
	"github.com/sp0x/torrentd/indexer/search"
	"google.golang.org/api/iterator"
)

//GetLatest returns the latest `count` of records.
func (f *FirestoreStorage) GetLatest(count int) []search.ResultItemBase {
	var output []search.ResultItemBase
	collection := f.getCollection()
	iter := collection.OrderBy("ID", firestore.Desc).Limit(count).Documents(f.context)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if doc == nil {
			continue
		}
		newItem := f.marshaler.New()
		err = doc.DataTo(newItem)
		if err != nil {
			continue
		}
		output = append(output, newItem.(search.ResultItemBase))
	}
	return output
}
