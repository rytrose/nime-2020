package main

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Timeouts (in seconds) for mongodb interactions
const (
	TimeoutConnect = 10
	TimeoutOp      = 2
)

// DB is a wrapper around a mongodb client.
type DB struct {
	client *mongo.Client
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

	return &DB{
		client: client,
	}
}
