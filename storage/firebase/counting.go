package firebase

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"strconv"
)

type counter struct {
	numOfShards int
}
type counterShard struct {
	Count int
}

func (c counter) initCounterIfNeeded(collection *firestore.CollectionRef, ctx context.Context, doc string) error {
	_, err := collection.Doc(doc).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		} else {
			//Doc does not exist, we'll create it
		}
	} else {
		return nil
	}
	return c.initCounter(ctx, collection.Doc(doc))
}

func (c *counter) initCounter(ctx context.Context, docRef *firestore.DocumentRef) error {
	colRef := docRef.Collection("shards")

	// Initialize each shard with count=0
	for num := 0; num < c.numOfShards; num++ {
		shard := counterShard{0}

		if _, err := colRef.Doc(strconv.Itoa(num)).Set(ctx, shard); err != nil {
			return fmt.Errorf("set: %v", err)
		}
	}
	return nil
}

//Increment the count
func (c *counter) incrementCounter(ctx context.Context, docRef *firestore.DocumentRef) (*firestore.WriteResult, error) {
	//Chose a random shard
	docID := strconv.Itoa(rand.Intn(c.numOfShards))
	//Get it
	shardRef := docRef.Collection("shards").Doc(docID)
	//Increment it
	return shardRef.Update(ctx, []firestore.Update{
		{Path: "Count", Value: firestore.Increment(1)},
	})
}

// getCount returns a total count across all shards.
func (c *counter) getCount(ctx context.Context, docRef *firestore.DocumentRef) (int64, error) {
	var total int64
	shards := docRef.Collection("shards").Documents(ctx)
	for {
		doc, err := shards.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("next: %v", err)
		}

		vTotal := doc.Data()["Count"]
		shardCount, ok := vTotal.(int64)
		if !ok {
			return 0, fmt.Errorf("firestore: invalid dataType %T, want int64", vTotal)
		}
		total += shardCount
	}
	return total, nil
}
