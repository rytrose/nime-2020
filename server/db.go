package main

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Timeouts (in seconds) for mongodb interactions
const (
	TimeoutConnect = 30
	TimeoutOp      = 10
)

// database is the common reference to mongo
var database *DB

// DB is a wrapper around a mongodb client.
type DB struct {
	client  *mongo.Client
	db      *mongo.Database
	roomCol *mongo.Collection
}

// NewDB creates a connection to the mongodb.
func NewDB(uri string) *DB {
	ctx, _ := context.WithTimeout(context.Background(), TimeoutConnect*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("DB connect error: %s", err)
	}
	ctx, _ = context.WithTimeout(context.Background(), TimeoutOp*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("DB ping error: %s", err)
	}

	// Reference database
	db := client.Database("nime2020")

	// Reference collections
	roomCol := db.Collection("room")

	return &DB{
		client:  client,
		db:      db,
		roomCol: roomCol,
	}
}

// GetState fetches room state from mongo, creating it if it does not exist.
func (db *DB) GetState(roomID string) (bson.M, error) {
	ctx, _ := context.WithTimeout(context.Background(), TimeoutOp*time.Second)
	var result bson.M
	err := db.roomCol.FindOne(ctx, bson.M{"roomID": roomID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// For now, create document when one does not exist
			// TODO: create document when syncing with firestore
			ctx, _ := context.WithTimeout(context.Background(), TimeoutOp*time.Second)
			doc := bson.M{"roomID": roomID, "test": 0}
			_, err = db.roomCol.InsertOne(ctx, doc)
			if err != nil {
				return nil, fmt.Errorf("database insert error: %s", err)
			}
			return doc, nil
		}
		return nil, fmt.Errorf("database find error: %s", err)
	}
	return result, nil
}

// CommitOperation gets room state from mongo and adds an operation.
func (db *DB) CommitOperation(roomID string) (bson.M, error) {
	doc, err := db.GetState(roomID)
	if err != nil {
		return nil, err
	}
	testVal, ok := doc["test"]
	if !ok {
		return nil, fmt.Errorf("didn't find test in record")
	}
	testV := testVal.(int32)

	ctx, _ := context.WithTimeout(context.Background(), TimeoutOp*time.Second)
	var result bson.M
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = db.roomCol.FindOneAndUpdate(ctx, bson.M{"roomID": roomID}, bson.D{{"$set", bson.D{{"test", testV + 1}}}}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("unable to find room state to update")
		}
		return nil, fmt.Errorf("database update error: %s", err)
	}
	return result, nil
}
