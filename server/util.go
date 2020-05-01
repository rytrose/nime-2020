package main

import (
	"math/rand"
	"sync"
)

// ClientMap is a concurrency-safe map of clients to bool (map for indexing).
type ClientMap struct {
	sync.RWMutex
	m map[*Client]bool
}

// NewClientMap instantiates a ClientMap.
func NewClientMap() *ClientMap {
	return &ClientMap{
		m: make(map[*Client]bool),
	}
}

// Get returns a value from the map.
func (c *ClientMap) Get(k *Client) (bool, bool) {
	c.RLock()
	v, ok := c.m[k]
	c.RUnlock()
	return v, ok
}

// Set sets a value in the map.
func (c *ClientMap) Set(k *Client, v bool) {
	c.Lock()
	c.m[k] = v
	c.Unlock()
}

// Delete deletes a key from the map.
func (c *ClientMap) Delete(k *Client) {
	c.Lock()
	delete(c.m, k)
	c.Unlock()
}

// Range iterates over the map.
func (c *ClientMap) Range(f func(k *Client, v bool) bool) {
	c.RLock()
	for k, v := range c.m {
		ok := f(k, v)
		if !ok {
			break
		}
	}
	c.RUnlock()
}

// GetRandomClient returns a random client from the map.
func (c *ClientMap) GetRandomClient() *Client {
	c.Lock()
	defer c.Unlock()
	length := len(c.m)
	if length == 0 {
		return nil
	}
	ind := rand.Intn(length)
	i := 0
	for client := range c.m {
		if i == ind {
			return client
		}
		i++
	}
	return nil
}

// RoomMap is a concurrency-safe map of room names to rooms.
type RoomMap struct {
	sync.RWMutex
	m map[string]*Room
}

// NewRoomMap instantiates a RoomMap.
func NewRoomMap() *RoomMap {
	return &RoomMap{
		m: make(map[string]*Room),
	}
}

// Get returns a value from the map.
func (r *RoomMap) Get(k string) (*Room, bool) {
	r.RLock()
	v, ok := r.m[k]
	r.RUnlock()
	return v, ok
}

// Set sets a value in the map.
func (r *RoomMap) Set(k string, v *Room) {
	r.Lock()
	r.m[k] = v
	r.Unlock()
}

// Delete deletes a key from the map.
func (r *RoomMap) Delete(k string) {
	r.Lock()
	delete(r.m, k)
	r.Unlock()
}
