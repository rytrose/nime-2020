package main

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Message types
const (
	TypeAnnounce    = "announce"
	TypeEnterRoom   = "enterRoom"
	TypeExitRoom    = "exitRoom"
	TypeOperation   = "operation"
	TypeStateUpdate = "stateUpdate"
)

// Message is the object websocket clients send.
type Message struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	UserID string `json:"userID"`
	RoomID string `json:"roomID"`
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
		log.Debugf("user \"%s\" announced", m.UserID)
		c.UserID = m.UserID
	case TypeEnterRoom:
		if err != nil {
			c.Send(bson.M{
				"id":    m.ID,
				"error": fmt.Sprintf("error getting state: %s", err),
			})
			break
		}
		room, ok := rooms[m.RoomID]
		if !ok {
			// For now create room
			rooms[m.RoomID] = &Room{
				RoomID:  m.RoomID,
				Members: make(map[*Client]bool),
			}
			room = rooms[m.RoomID]
			// c.Send(bson.M{
			// 	"id":    m.ID,
			// 	"error": fmt.Sprintf("room %s does not exist (TODO: sync from firestore)", m.RoomID),
			// })
		}
		room.Members[c] = true
		c.Room = room
		doc, err := database.GetState(m.RoomID)
		if err != nil {
			c.Send(bson.M{
				"id":    m.ID,
				"error": err,
			})
			break
		}
		c.Send(bson.M{
			"id":       m.ID,
			"roomData": doc,
		})
	case TypeExitRoom:
		if c.Room == nil {
			c.Send(bson.M{
				"id":    m.ID,
				"error": fmt.Sprintf("user %s is not in a room to exit", c.UserID),
			})
			break
		}
		delete(c.Room.Members, c)
		c.Room = nil
		c.Send(bson.M{
			"id": m.ID,
		})
	case TypeOperation:
		if c.Room == nil {
			c.Send(bson.M{
				"error": fmt.Sprintf("user %s is not in a room to commit an operation", c.UserID),
			})
			break
		}
		doc, err := database.CommitOperation(c.Room.RoomID)
		if err != nil {
			c.Send(bson.M{
				"error": err,
			})
			break
		}
		c.Room.Broadcast(bson.M{
			"type":  TypeStateUpdate,
			"state": doc,
		})
	default:
		log.Warnf("message type \"%s\" not implemented", m.Type)
	}
}
