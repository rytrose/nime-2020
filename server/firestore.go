package main

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// Timeouts (in seconds) for mongodb interactions
const (
	FSTimeoutOp = 2
)

// fs is the common reference to mongo
var fs *Firestore

// Firestore is a wrapper around a firestore client
type Firestore struct {
	client  *firestore.Client
	roomCol *firestore.CollectionRef
}

// NewFirestore creates a firestore client.
func NewFirestore() *Firestore {
	ctx, _ := context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	sa := option.WithCredentialsFile("fir-test-9a9f3-firebase-adminsdk-hbzpi-90259f5885.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to create firebase app: %s", err))
	}

	ctx, _ = context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to create firestore client: %s", err))
	}

	return &Firestore{
		client:  client,
		roomCol: client.Collection("rooms"),
	}
}

// GetRoom retrieves a room from firestore.
func (fs *Firestore) GetRoom(roomName string) (bson.M, error) {
	ctx, _ := context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	doc, err := fs.roomCol.Doc(roomName).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get room from firestore: %s", err)
	}
	return doc.Data(), nil
}
