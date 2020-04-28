package main

// Adapted from https://gitRoom.com/gorilla/websocket/tree/master/examples/chat

import (
	log "github.com/sirupsen/logrus"
)

// rooms contain all the existing rooms
var rooms = map[string]*Room{}

// Room maintains the set of active members and broadcasts messages to the room members.
type Room struct {
	// RoomName is the name for the room.
	RoomName string

	// Members are registered room members.
	Members map[*Client]bool
}

// Broadcast sends a message to all connected members.
func (r *Room) Broadcast(m interface{}) {
	for c := range r.Members {
		err := c.Send(m)
		if err != nil {
			log.Errorf("%s", err)
		}
	}
}
