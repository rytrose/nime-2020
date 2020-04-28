package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Announce registers a user with a client connection.
func Announce(c *Client, m *Message) {
	log.Debugf("user \"%s\" announced", m.UserID)
	c.UserID = m.UserID
}

// EnterRoom registers a client with a room.
func EnterRoom(c *Client, m *Message) bson.M {
	// Get room
	room, ok := rooms[m.RoomName]
	if !ok {
		// For now create room, TODO: fetch rooms from firestore
		rooms[m.RoomName] = &Room{
			RoomName: m.RoomName,
			Members:  make(map[*Client]bool),
		}
		room = rooms[m.RoomName]
	}

	// Add client to room
	room.Members[c] = true
	c.Room = room

	// Get room data
	doc, err := database.GetRoom(m.RoomName)
	if err != nil {
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("unable to get room: %s", err),
		}
	}
	ops, err := database.GetAllOperations(m.RoomName)
	if err != nil {
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("unable to get all operations: %s", err),
		}
	}
	return bson.M{
		"id":         m.ID,
		"roomDoc":    doc,
		"operations": ops,
	}
}

// ExitRoom unregisters a client from a room.
func ExitRoom(c *Client, m *Message) bson.M {
	if c.Room == nil {
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("user %s is not in a room to exit", c.UserID),
		}
	}
	delete(c.Room.Members, c)
	c.Room = nil
	return bson.M{
		"id": m.ID,
	}
}

// Operate commits an operation to a room.
func Operate(c *Client, m *Message) (bson.M, bson.M) {
	if c.Room == nil {
		return nil, bson.M{
			"error": fmt.Sprintf("user %s is not in a room to commit an operation", c.UserID),
		}
	}
	doc, err := database.CommitOperation(c.Room.RoomName, m.Operation)
	if err != nil {
		return nil, bson.M{
			"error": fmt.Sprintf("unable to commit operation: %s", err),
		}
	}
	return bson.M{
		"type":      TypeOperationUpdate,
		"operation": doc.Ops[len(doc.Ops)-1],
	}, nil
}
