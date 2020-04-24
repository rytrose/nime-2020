package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// Message types
const (
	TypeEnterRoom = "enterRoom"
	TypeExitRoom  = "exitRoom"
	TypeGetState  = "getState"
)

// Message is the object websocket clients send.
type Message struct {
	Type   string `json:"type"`
	RoomID string `json:"roomID"`
}

// dispatch fans out different types of messages from websocket clients.
func dispatch(c *Client, b []byte) {
	m := &Message{}
	err := json.Unmarshal(b, m)
	if err != nil {
		log.Errorf("unable to unmarshal message: %s", err)
	}

	switch m.Type {
	case TypeEnterRoom:
	case TypeGetState:
		if c.room == nil {
			log.Warnf("client is not in a room to getState")
			return
		}
		c.room.GetState(m)
	default:
		log.Warnf("message type \"%s\" not implemented", m.Type)
	}
}
