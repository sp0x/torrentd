package firebase

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type counter struct {
	numOfShards int
}

type counterShard struct {
	Count int
}

func (c counter) initCounterIfNeeded(ctx context.Context, collection *firestore.CollectionRef, doc string) error {
	_, err := collection.Doc(doc).Get(ctx)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}
	} else {
		return nil
	}
	return c.initCounter(ctx, collection.Doc(doc))
}

func (c *counter) initCounter(ctx context.Context, docRef *firestore.DocumentRef) error {
	counterDoc := make(map[string]interface{})
	counterDoc["created"] = time.Now()
	_, err := docRef.Create(ctx, counterDoc)
	if err != nil {
		return err
	}
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

// Increment the count
func (c *counter) incrementCounter(ctx context.Context, docRef *firestore.DocumentRef) (*firestore.WriteResult, error) {
	// Chose a random shard
	docID := strconv.Itoa(rand.Intn(c.numOfShards))
	// Get it
	shardRef := docRef.Collection("shards").Doc(docID)
	// Increment it
	return shardRef.Update(ctx, []firestore.Update{
		{Path: "Count", Value: firestore.Increment(1)},
	})
}
