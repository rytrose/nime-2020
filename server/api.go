package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Message types
const (
	TypeAnnounce        = "announce"        // [Client->Server] Provides a user ID to the client
	TypeEnterRoom       = "enterRoom"       // [Client->Server] Client associates with a room, and requests the current state
	TypeExitRoom        = "exitRoom"        // [Client->Server] Client disassociates with a room
	TypeOperation       = "operation"       // [Client->Server] Client makes and operation
	TypeOperationUpdate = "operationUpdate" // [Server->Client] Server disseminates an operation to all Clients in a room
	TypeRequestState    = "requestState"    // [Server->Client] Server asks a Client for the full state of the room
	TypeState           = "state"           // [Client->Server] Client sends the full state to the server
	TypeClearState      = "clearState"      // [Server->Client] Server tells a Client to clear the current state
)

// Message is the superset of the object websocket clients send.
type Message struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	UserID        string `json:"userID"`
	RoomName      string `json:"roomName"`
	OperationType string `json:"operationType"`
	Operation     bson.M `json:"operation"`
	State         bson.M `json:"state"`
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
		Announce(c, m)
	case TypeEnterRoom:
		res := EnterRoom(c, m)
		c.Send(res)
	case TypeExitRoom:
		res := ExitRoom(c, m)
		c.Send(res)
	case TypeOperation:
		res, err := Operate(c, m)
		if err != nil {
			c.Send(err)
			break
		}
		c.Room.Broadcast(res, c)
	case TypeState:
		State(c, m)
	default:
		log.Warnf("message type \"%s\" not implemented", m.Type)
	}
}
