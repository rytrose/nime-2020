package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// AnnounceHandler registers a user with a client connection.
func AnnounceHandler(c *Client, m *Message) bson.M {
	ok := true
	f := func(client *Client, _ bool) bool {
		if client.UserID == m.UserID {
			ok = false
			return false
		}
		return true
	}
	clients.Range(f)

	if !ok {
		c.UserID = m.UserID
		log.Warnf("(WARNING: multiple instances of the same user) user \"%s\" announced", m.UserID)
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("user %s already connected", m.UserID),
		}
	}

	c.UserID = m.UserID
	log.Debugf("user \"%s\" announced", m.UserID)

	return bson.M{
		"id": m.ID,
	}
}

// EnterRoomHandler registers a client with a room.
func EnterRoomHandler(c *Client, m *Message) bson.M {
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

	// Get room data (creates from firestore if doesn't exist)
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

	// Increment room num_members
	doc, err = database.UpdateRoomNumMembers(m.RoomName, 1)
	if err != nil {
		c.Room = nil
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("unable to increment room num_members: %s", err),
		}
	}

	// Update clients with num_members
	c.Room.Broadcast(bson.M{
		"type":       TypeNumMembersUpdate,
		"numMembers": doc.NumMembers,
	}, c)

	// Add client to room
	room.Members.Set(c, true)

	return bson.M{
		"id":         m.ID,
		"roomDoc":    doc,
		"state":      state,
		"operations": operations,
	}
}

// ExitRoomHandler unregisters a client from a room.
func ExitRoomHandler(c *Client, m *Message) bson.M {
	// Check if client is in room
	if c.Room == nil {
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("user %s is not in a room to exit", c.UserID),
		}
	}

	// Decrement room num_members
	doc, err := database.UpdateRoomNumMembers(m.RoomName, -1)
	if err != nil {
		return bson.M{
			"id":    m.ID,
			"error": fmt.Sprintf("unable to decrement room num_members: %s", err),
		}
	}

	// Update clients with num_members
	c.Room.Broadcast(bson.M{
		"type":       TypeNumMembersUpdate,
		"numMembers": doc.NumMembers,
	}, c)

	c.Room.Members.Delete(c)
	c.Room = nil
	return bson.M{
		"id": m.ID,
	}
}

// OperationsHandler commits operations to a room.
func OperationsHandler(c *Client, m *Message) (bson.M, bson.M) {
	if c.Room == nil {
		return nil, bson.M{
			"error": fmt.Sprintf("user %s is not in a room to commit operations", c.UserID),
		}
	}
	ops, err := database.CommitOperations(c.Room.RoomName, m.Operations)
	if err != nil {
		return nil, bson.M{
			"error": fmt.Sprintf("unable to commit operation: %s", err),
		}
	}
	return bson.M{
		"type":       TypeOperationsUpdate,
		"operations": ops,
	}, nil
}

// StateHandler receives the full state from a client in order to send to other clients who need it.
func StateHandler(c *Client, m *Message) {
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
