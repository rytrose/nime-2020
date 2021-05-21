package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Timeouts (in seconds) for mongodb interactions
const (
	DBTimeoutConnect = 10
	DBTimeoutOp      = 2
	MaxOpsPerBucket  = 100
)

// database is the common reference to mongo
var database *DB

// DB is a wrapper around a mongodb client.
type DB struct {
	client              *mongo.Client
	db                  *mongo.Database
	roomCol             *mongo.Collection
	operationBucketsCol *mongo.Collection
	maxOpsPerBucket     int
	writeMutex          sync.Mutex
}

// RoomDoc is a document that stores metadata about a room.
type RoomDoc struct {
	ID         primitive.ObjectID `bson:"_id"`
	RoomName   string             `bson:"room_name"`
	NumBuckets int                `bson:"num_buckets"`
	NumMembers int                `bson:"num_members"`
}

// OpBucketDoc is a document that stores operations.
type OpBucketDoc struct {
	ID       primitive.ObjectID `bson:"_id"`
	RoomName string             `bson:"room_name"`
	Bucket   int                `bson:"bucket"`
	Count    int                `bson:"count"`
	Ops      []bson.M           `bson:"operations"`
}

// NewDB creates a connection to the mongodb.
func NewDB(uri string) *DB {
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutConnect*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("DB connect error: %s", err)
	}
	ctx, _ = context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("DB ping error: %s", err)
	}

	// Reference database
	db := client.Database("nime2020")

	// Reference collections
	// Adapted hybrid comments pattern: https://docs.mongodb.com/drivers/use-cases/storing-comments
	roomCol := db.Collection("room")
	operationBucketsCol := db.Collection("operationBuckets")

	dbObj := &DB{
		client:              client,
		db:                  db,
		roomCol:             roomCol,
		operationBucketsCol: operationBucketsCol,
		maxOpsPerBucket:     MaxOpsPerBucket,
	}

	// Ensure indicies
	dbObj.configureIndices()
	return dbObj
}

// configureIndices ensure the DB has the necessary indices created.
func (db *DB) configureIndices() {
	// Index names to ensure exist
	roomNameIndexName := "room_name"
	opBucketIndexName := "room_name_bucket"
	expectedIndices := map[string]bool{
		roomNameIndexName: false,
		opBucketIndexName: false,
	}

	// List indices - ROOM
	opts := options.ListIndexes().SetMaxTime(DBTimeoutOp * time.Second)
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	cursor, err := db.roomCol.Indexes().List(ctx, opts)
	if err != nil {
		log.Fatalf("DB index list error: %s", err)
	}
	var roomIndRes []bson.M
	if err = cursor.All(context.Background(), &roomIndRes); err != nil {
		log.Fatalf("DB index list cursor error: %s", err)
	}

	// Check if known indices are created
	for _, ind := range roomIndRes {
		name := ind["name"].(string)
		log.Infof("existing index: %+v", ind)
		_, ok := expectedIndices[name]
		if ok {
			expectedIndices[name] = true
		}
	}

	// List indices - OP BUCKETS
	opts = options.ListIndexes().SetMaxTime(DBTimeoutOp * time.Second)
	ctx, _ = context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	cursor, err = db.operationBucketsCol.Indexes().List(ctx, opts)
	if err != nil {
		log.Fatalf("DB index list error: %s", err)
	}
	var opBucketsIndRes []bson.M
	if err = cursor.All(context.Background(), &opBucketsIndRes); err != nil {
		log.Fatalf("DB index list cursor error: %s", err)
	}

	// Check if known indices are created
	for _, ind := range opBucketsIndRes {
		name := ind["name"].(string)
		log.Infof("existing index: %+v", ind)
		_, ok := expectedIndices[name]
		if ok {
			expectedIndices[name] = true
		}
	}

	// Create indices that don't yet exist
	for indexName, created := range expectedIndices {
		if !created {
			switch indexName {
			case roomNameIndexName:
				roomIdxModel := mongo.IndexModel{
					Keys: bson.M{
						"room_name": 1,
					},
					Options: options.Index().SetName(roomNameIndexName),
				}
				ctx, _ = context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
				_, err = db.roomCol.Indexes().CreateOne(ctx, roomIdxModel)
				if err != nil {
					log.Fatalf("unable to ensure room index: %s", err)
				}
				break
			case opBucketIndexName:
				operationBucketsIdxModel := mongo.IndexModel{
					Keys: bson.M{
						"room_name": 1,
						"bucket":    1,
					},
					Options: options.Index().SetName(opBucketIndexName),
				}
				ctx, _ = context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
				_, err = db.operationBucketsCol.Indexes().CreateOne(ctx, operationBucketsIdxModel)
				if err != nil {
					log.Fatalf("unable to ensure op bucket index: %s", err)
				}
				break
			}
			log.Infof("created index %s", indexName)
		}
	}
}

// GetRoom gets the room document given a human-readable roomName.
func (db *DB) GetRoom(roomName string) (*RoomDoc, error) {
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	query := bson.M{"room_name": roomName}

	room := &RoomDoc{}
	err := db.roomCol.FindOne(ctx, query).Decode(room)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Look up room in firestore
			roomDoc, err := fb.GetRoom(roomName)
			if err != nil {
				return nil, err
			}

			// Create room in mongo
			log.Debugf("creating room from firebase: %s", roomDoc)
			ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
			room = &RoomDoc{
				ID:         primitive.NewObjectID(),
				RoomName:   roomName,
				NumBuckets: 1,
				NumMembers: 0,
			}
			res, err := db.roomCol.InsertOne(ctx, room)
			if err != nil {
				return nil, fmt.Errorf("database insert error: %s", err)
			}
			room.ID = res.InsertedID.(primitive.ObjectID)
			return room, nil
		}
		return nil, fmt.Errorf("database find error: %s", err)
	}
	return room, nil
}

// UpdateRoomNumMembers increments/decrements the number of members in a room.
func (db *DB) UpdateRoomNumMembers(roomName string, updateIncrement int) (*RoomDoc, error) {
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	query := bson.M{"room_name": roomName}
	operation := bson.M{"$inc": bson.M{"num_members": updateIncrement}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	roomDoc := &RoomDoc{}
	err := db.roomCol.FindOneAndUpdate(ctx, query, operation, opts).Decode(roomDoc)
	if err != nil {
		return nil, fmt.Errorf("database update room num_members error: %s", err)
	}
	return roomDoc, nil
}

// commitOperation stores an operation committed in a room.
func (db *DB) commitOperation(roomDoc *RoomDoc, op bson.M) (*OpBucketDoc, error) {
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	query := bson.M{"room_name": roomDoc.RoomName, "bucket": roomDoc.NumBuckets}
	operation := bson.M{"$inc": bson.M{"count": 1}, "$push": bson.M{"operations": op}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	opBucket := &OpBucketDoc{}
	err := db.operationBucketsCol.FindOneAndUpdate(ctx, query, operation, opts).Decode(opBucket)
	if err != nil {
		return nil, fmt.Errorf("database update op bucket with op error: %s", err)
	}

	if opBucket.Count == db.maxOpsPerBucket {
		ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
		query := bson.M{"_id": roomDoc.ID, "num_buckets": roomDoc.NumBuckets}
		update := bson.M{"$inc": bson.M{"num_buckets": 1}}

		_, err = db.roomCol.UpdateOne(ctx, query, update)
		if err != nil {
			return nil, fmt.Errorf("database update num op buckets error: %s", err)
		}
	}

	return opBucket, nil
}

// CommitOperations writes operations committed in a room.
func (db *DB) CommitOperations(roomName string, ops []bson.M) ([]bson.M, error) {
	room, err := db.GetRoom(roomName)
	if err != nil {
		return nil, fmt.Errorf("unable to get room: %w", err)
	}

	// Ensure all operations submitted together are written together
	db.writeMutex.Lock()
	defer db.writeMutex.Unlock()

	// Commit all operations
	for _, op := range ops {
		_, err := db.commitOperation(room, op)
		if err != nil {
			return nil, fmt.Errorf("unable to commit operation: %w", err)
		}
	}

	return ops, nil
}

// GetAllOperations returns the full history of operations for a given room.
func (db *DB) GetAllOperations(roomName string) ([]bson.M, error) {
	all := []bson.M{}
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	query := bson.M{"room_name": roomName}

	cursor, err := db.operationBucketsCol.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("database find error: %s", err)
	}
	var results []OpBucketDoc
	ctx, _ = context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("database find cursor error: %s", err)
	}
	for _, bucketDoc := range results {
		all = append(all, bucketDoc.Ops...)
	}
	return all, nil
}

// DeleteAllOperations deletes all operations for a given room.
func (db *DB) DeleteAllOperations(roomName string) error {
	// Delete all buckets
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	query := bson.M{"room_name": roomName}

	_, err := db.operationBucketsCol.DeleteMany(ctx, query)
	if err != nil {
		return fmt.Errorf("database delete many error: %s", err)
	}

	// Set num_buckets to one
	ctx, _ = context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	query = bson.M{"room_name": roomName}
	operation := bson.M{"$set": bson.M{"num_buckets": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	roomDoc := &RoomDoc{}
	err = db.roomCol.FindOneAndUpdate(ctx, query, operation, opts).Decode(roomDoc)
	if err != nil {
		return fmt.Errorf("database update room num_buckets error: %s", err)
	}
	if roomDoc.NumBuckets != 1 {
		return fmt.Errorf("num_buckets of room %s was not set to 1 after deleting all operations", roomName)
	}

	// Reset all clients
	room, ok := rooms.Get(roomName)
	if !ok {
		return fmt.Errorf("server is not tracking room %s, but its operations have been deleted", roomName)
	}
	room.Broadcast(bson.M{
		"type": TypeClearState,
	})

	return nil
}

// ResetNumMembers sets the numMembers to 0 for all rooms.
func (db *DB) ResetNumMembers() error {
	// Update all NumMembers to 0
	ctx, _ := context.WithTimeout(context.Background(), DBTimeoutOp*time.Second)
	filter := bson.M{}
	update := bson.M{"$set": bson.M{"num_members": 0}}

	updateResult, err := db.roomCol.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("database update many error: %s", err)
	}
	log.Infof("reset NumMembers for %d (filter matched %d)", updateResult.ModifiedCount, updateResult.MatchedCount)

	return nil
}
