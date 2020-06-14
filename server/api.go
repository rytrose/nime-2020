package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Message types
const (
	TypeAnnounce   = "announce"   // [Client->Server] Provides a user ID to the client
	TypeEnterRoom  = "enterRoom"  // [Client->Server] Client associates with a room, and requests the current state
	TypeExitRoom   = "exitRoom"   // [Client->Server] Client disassociates with a room
	TypeOperations = "operations" // [Client->Server] Client makes submits operations
	TypeState      = "state"      // [Client->Server] Client sends the full state to the server

	TypeOperationsUpdate = "operationsUpdate" // [Server->Client] Server disseminates operations to all Clients in a room
	TypeRequestState     = "requestState"     // [Server->Client] Server asks a Client for the full state of the room
	TypeClearState       = "clearState"       // [Server->Client] Server tells a Client to clear the current state
	TypeNumMembersUpdate = "numMembersUpdate" // [Server->Client] Server tells a Client how many members are in the room
)

// Message is the superset of the object websocket clients send.
type Message struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	UserID        string   `json:"userID"`
	RoomName      string   `json:"roomName"`
	OperationType string   `json:"operationType"`
	Operations    []bson.M `json:"operations"`
	State         bson.M   `json:"state"`
}

// dispatch fans out different types of messages from websocket clients.
func dispatch(c *Client, b []byte) {
	m := &Message{}
	err := json.Unmarshal(b, m)
	if err != nil {
		log.Errorf("unable to unmarshal message (%s): %s", b, err)
	}

	switch m.Type {
	case TypeAnnounce:
		res := AnnounceHandler(c, m)
		c.Send(res)
	case TypeEnterRoom:
		res := EnterRoomHandler(c, m)
		c.Send(res)
	case TypeExitRoom:
		res := ExitRoomHandler(c, m)
		c.Send(res)
	case TypeOperations:
		res, err := OperationsHandler(c, m)
		if err != nil {
			c.Send(err)
			break
		}
		c.Room.Broadcast(res, c)
	case TypeState:
		StateHandler(c, m)
	default:
		log.Warnf("message type \"%s\" not implemented", m.Type)
	}
}
