package main

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Timeouts (in seconds) for mongodb interactions
const (
	FSTimeoutOp = 2
)

// fb is the common reference to firebase
var fb *Firebase

// Firebase is a wrapper around a firebase client
type Firebase struct {
	firestoreClient *firestore.Client
	authClient      *auth.Client
	roomCol         *firestore.CollectionRef
}

// NewFirebase creates a firebase client.
func NewFirebase() *Firebase {
	ctx, _ := context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	sa := option.WithCredentialsFile("fir-test-9a9f3-firebase-adminsdk-hbzpi-90259f5885.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to create firebase app: %s", err))
	}

	ctx, _ = context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to create firestore client: %s", err))
	}

	ctx, _ = context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to create auth client: %s", err))
	}

	return &Firebase{
		firestoreClient: firestoreClient,
		authClient:      authClient,
		roomCol:         firestoreClient.Collection("rooms"),
	}
}

// GetRoom retrieves a room from firestore.
func (fb *Firebase) GetRoom(roomName string) (bson.M, error) {
	ctx, _ := context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	doc, err := fb.roomCol.Doc(roomName).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get room from firestore: %s", err)
	}
	return doc.Data(), nil
}

// DeleteAllUsers deletes all users from firebase. Be careful.
func (fb *Firebase) DeleteAllUsers() error {
	// Get all users
	ctx, _ := context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
	iter := fb.authClient.Users(ctx, "")
	for {
		// Get the next user
		user, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("error listing users: %s", err)
		}

		// Delete user
		log.Debugf("deleting user %s", user.UID)
		ctx, _ := context.WithTimeout(context.Background(), FSTimeoutOp*time.Second)
		err = fb.authClient.DeleteUser(ctx, user.UID)
		if err != nil {
			log.Fatalf("unable to delete user %s: %s", user.UID, err)
		}
	}

	return nil
}
