package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"

	log "github.com/sirupsen/logrus"
)

// clients contains all existing clients.
var clients = NewClientMap()

// Client is a wrapper around the websocket connection
type Client struct {
	// Connection ID for the client.
	connID string

	// User ID for the client.
	UserID string

	// The room the client is a member of.
	Room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Timeout for channel operations in milliseconds.
	chanTimeout int

	// Buffered channel of outbound messages.
	send chan interface{}

	// Flag for whether the send chan is open
	sendOpen bool

	// Channel to wait on for full state update.
	stateUpdate chan bson.M
}

// NewClient creates and starts a new Client.
func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		connID:      uuid.New().String(),
		UserID:      "", // To be populated on TypeAnnounce
		Room:        nil,
		conn:        conn,
		chanTimeout: 500,
		send:        make(chan interface{}),
		sendOpen:    true,
		stateUpdate: make(chan bson.M),
	}
	go c.reader()
	go c.writer()
	clients.Set(c, true)
	return c
}

// Close frees up the websocket and removes it from memory.
func (c *Client) Close() {
	log.Infof("closing connection %s", c.connID)

	// Close send chan
	if c.sendOpen {
		c.sendOpen = false
		close(c.send)
	}

	// Clean up room presence
	if c.Room != nil {
		// Decrement room num_members
		doc, err := database.UpdateRoomNumMembers(c.Room.RoomName, -1)
		if err != nil {
			log.Errorf("[DATA OUT OF SYNC] unable to decrement num_members for room %s: %s", c.Room.RoomName, err)
		} else {
			// Update clients with num_members
			c.Room.Broadcast(bson.M{
				"type":       TypeNumMembersUpdate,
				"numMembers": doc.NumMembers,
			}, c)
		}
		c.Room.Members.Delete(c)
		c.Room = nil
	}
	clients.Delete(c)
}

// Send sends a message to the connected websocket client.
func (c *Client) Send(v interface{}) error {
	if c.sendOpen {
		select {
		case c.send <- v:
			return nil
		case <-time.After(time.Duration(c.chanTimeout) * time.Millisecond):
			return fmt.Errorf("unable to send message (send channel timeout)")
		}
	}
	return fmt.Errorf("attempted send on closed send channel")
}

// WaitForState waits for the full state to be provided to the client from another.
func (c *Client) WaitForState() (bson.M, error) {
	if c.Room == nil {
		return nil, fmt.Errorf("client not in room to receive state")
	}
	c.Room.NeedsState.Set(c, true)
	defer func() {
		c.Room.NeedsState.Delete(c)
	}()

	select {
	case state := <-c.stateUpdate:
		return state, nil
	case <-time.After(time.Duration(c.chanTimeout) * time.Millisecond):
		return nil, fmt.Errorf("did not receive full state (receive channel timeout)")
	}
}

// ReceiveState receives the full state of a room.
func (c *Client) ReceiveState(state bson.M) {
	select {
	case c.stateUpdate <- state:
	case <-time.After(time.Duration(c.chanTimeout) * time.Millisecond):
		log.Errorf("unable to send state (send channel timeout)")
	}
}

// reader loops over and dispatches incoming messages.
func (c *Client) reader() {
	for {
		_, m, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("unexpected close error: %v", err)
			}
			c.Close()
			break
		}
		log.Debugf("received message: %s", m)
		go dispatch(c, m)
	}
}

// writer loops over the send channel and sends messages.
func (c *Client) writer() {
	for {
		m, ok := <-c.send
		if !ok {
			// Send channel has been closed
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		// Write message as JSON
		err := c.conn.WriteJSON(m)
		if err != nil {
			if err == websocket.ErrCloseSent {
				// Don't log error on closed channel
			}
			log.Errorf("error writing message: %s", err)
		}
	}
}
