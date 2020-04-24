package main

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

// clients contains all existing clients
var clients = map[string]*Client{}

// Client is a wrapper around the websocket connection
type Client struct {
	id string

	// The room the client is a member of
	room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan interface{}
}

// NewClient creates and starts new Client
func NewClient(id string, conn *websocket.Conn) *Client {
	c := &Client{
		id:   id,
		room: nil,
		conn: conn,
		send: make(chan interface{}, 256),
	}
	go c.reader()
	go c.writer()
	clients[c.id] = c
	return c
}

// Close frees up the websocket and removes from memory
func (c *Client) Close() {
	c.conn.Close()
	delete(clients, c.id)
}

// Send sends a message to the connected websocket client
func (c *Client) Send(v interface{}) error {
	select {
	case c.send <- v:
		return nil
	case <-time.After(500 * time.Millisecond):
		return fmt.Errorf("unable to send message (send channel timeout)")
	}
}

// reader loops over and dispatches incoming messages
func (c *Client) reader() {
	defer func() {
		c.conn.Close()
	}()
	for {
		_, m, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			break
		}
		log.Infof("Received message: %s", m)
		go dispatch(c, m)
	}
}

// writer loops over the send channel and sends messages
func (c *Client) writer() {
	for {
		// Send buffered messages
		n := len(c.send)
		for i := 0; i < n; i++ {
			m, ok := <-c.send
			if !ok {
				// Send channel has been closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message as JSON
			err := c.conn.WriteJSON(m)
			if err != nil {
				log.Errorf("error writing message: %s", err)
			}
		}
	}
}
