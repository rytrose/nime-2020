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
	room, ok := rooms.Get(m.RoomName)
	if !ok {
		// Create Room object
		room = &Room{
			RoomName:   m.RoomName,
			Members:    NewClientMap(),
			NeedsState: NewClientMap(),
		}
		rooms.Set(m.RoomName, room)
	}

	// Get room data
	doc, err := database.GetRoom(m.RoomName)
	if err != nil {
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("unable to get room: %s", err),
		}
	}

	// Add room reference to client
	// 	but do not add client to room yet
	c.Room = room

	// Attempt to get room state
	var state bson.M
	var operations []bson.M
	existingClient := room.Members.GetRandomClient()
	if existingClient != nil {
		// Request full state from existing client in room
		existingClient.Send(bson.M{
			"type": TypeRequestState,
		})
		state, err = c.WaitForState()
		if err != nil {
			log.Warnf("unable to get full state: %s", err)
		}
	}
	if state == nil {
		// Fall back to sending all operations
		operations, err = database.GetAllOperations(m.RoomName)
		if err != nil {
			c.Room = nil
			return bson.M{
				"id":    m.ID,
				"error": fmt.Sprintf("unable to get full state or all operations: %s", err),
			}
		}
	}

	// Add client to room
	room.Members.Set(c, true)

	return bson.M{
		"id":         m.ID,
		"roomDoc":    doc,
		"state":      state,
		"operations": operations,
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
	c.Room.Members.Delete(c)
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

// State receives the full state from a client in order to send to other clients who need it.
func State(c *Client, m *Message) {
	room, ok := rooms.Get(m.RoomName)
	if !ok {
		log.Warnf("room %s doesn't exist", m.RoomName)
	}
	f := func(clientToUpdate *Client, _ bool) bool {
		// Don't block in order to not affect individual timeouts
		go clientToUpdate.ReceiveState(m.State)
		return true
	}
	room.NeedsState.Range(f)
}
