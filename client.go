package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *Client) WriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.WriteMessage(messageType, data)
	if err != nil {
		return err
	}

	return nil
}
