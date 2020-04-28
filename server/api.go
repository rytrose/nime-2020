package main

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Message types
const (
	TypeAnnounce        = "announce"
	TypeEnterRoom       = "enterRoom"
	TypeExitRoom        = "exitRoom"
	TypeOperation       = "operation"
	TypeOperationUpdate = "operationUpdate"
)

// Message is the object websocket clients send.
type Message struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	UserID        string `json:"userID"`
	RoomName      string `json:"roomName"`
	OperationType string `json:"operationType"`
	Operation     bson.M `json:"operation"`
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
		room, ok := rooms[m.RoomName]
		if !ok {
			// For now create room
			rooms[m.RoomName] = &Room{
				RoomName: m.RoomName,
				Members:  make(map[*Client]bool),
			}
			room = rooms[m.RoomName]
			// c.Send(bson.M{
			// 	"id":    m.ID,
			// 	"error": fmt.Sprintf("room %s does not exist (TODO: sync from firestore)", m.RoomName),
			// })
		}
		room.Members[c] = true
		c.Room = room
		doc, err := database.GetRoom(m.RoomName)
		if err != nil {
			log.Error(err)
			c.Send(bson.M{
				"id":    m.ID,
				"error": err,
			})
			break
		}
		ops, err := database.GetAllOperations(m.RoomName)
		if err != nil {
			log.Error(err)
			c.Send(bson.M{
				"id":    m.ID,
				"error": err,
			})
			break
		}
		c.Send(bson.M{
			"id":         m.ID,
			"roomDoc":    doc,
			"operations": ops,
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
		doc, err := database.CommitOperation(c.Room.RoomName, m.Operation)
		if err != nil {
			c.Send(bson.M{
				"error": err,
			})
			break
		}
		c.Room.Broadcast(bson.M{
			"type":      TypeOperationUpdate,
			"operation": doc.Ops[len(doc.Ops)-1],
		})
	default:
		log.Warnf("message type \"%s\" not implemented", m.Type)
	}
}
