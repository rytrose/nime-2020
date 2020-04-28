package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

// clients contains all existing clients.
var clients = map[string]*Client{}

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

	// Buffered channel of outbound messages.
	send chan interface{}
}

// NewClient creates and starts a new Client.
func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		connID: uuid.New().String(),
		UserID: "", // To be populated on TypeAnnounce
		Room:   nil,
		conn:   conn,
		send:   make(chan interface{}),
	}
	go c.reader()
	go c.writer()
	clients[c.connID] = c
	return c
}

// Close frees up the websocket and removes from memory.
func (c *Client) Close() {
	log.Infof("closing connection %s", c.connID)
	c.conn.Close()
	if c.Room != nil {
		delete(c.Room.Members, c)
	}
	delete(clients, c.connID)
}

// Send sends a message to the connected websocket client.
func (c *Client) Send(v interface{}) error {
	select {
	case c.send <- v:
		return nil
	case <-time.After(500 * time.Millisecond):
		return fmt.Errorf("unable to send message (send channel timeout)")
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
				log.Errorf("error writing message: %s", err)
			}
			log.Errorf("error writing message: %s", err)
		}
	}
}
