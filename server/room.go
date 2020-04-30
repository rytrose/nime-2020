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

// Broadcast sends a message to all connected members, except those passed in to ignore.
func (r *Room) Broadcast(m interface{}, ignoreClients ...*Client) {
	for c := range r.Members {
		ignore := false
		for _, toIgnore := range ignoreClients {
			if c == toIgnore {
				ignore = true
				break
			}
		}
		if !ignore {
			err := c.Send(m)
			if err != nil {
				log.Errorf("%s", err)
			}
		}
	}
}
