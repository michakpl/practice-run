package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	mu      sync.Mutex
	clients map[*Client]bool
}

type Chat struct {
	mu    sync.Mutex
	rooms map[string]Room
}

func (c *Chat) CreateRoom(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.rooms[name]; ok {
		return fmt.Errorf("room '%s' already exists", name)
	}

	c.rooms[name] = Room{
		clients: make(map[*Client]bool),
	}

	return nil
}

func (c *Chat) JoinRoom(client *Client, name string) (*Room, error) {
	room, ok := c.rooms[name]
	if !ok {
		return nil, fmt.Errorf("room '%s' do not exists", name)
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()

	if _, ok := room.clients[client]; ok {
		return nil, fmt.Errorf("client already exists")
	}

	room.clients[client] = true

	return &room, nil
}

func (c *Chat) LeaveRoom(client *Client, name string) error {
	room, ok := c.rooms[name]
	if !ok {
		return fmt.Errorf("room '%s' do not exists", name)
	}
	room.mu.Lock()
	defer room.mu.Unlock()

	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
	}

	return nil
}

func (c *Chat) GetRoom(name string) (*Room, error) {
	room, ok := c.rooms[name]
	if !ok {
		return nil, fmt.Errorf("room '%s' do not exists", name)
	}
	room.mu.Lock()
	defer room.mu.Unlock()

	return &room, nil
}

func (r *Room) SendMessage(username string, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for client := range r.clients {
		resMessage := Response{
			Action:   "receivedMessage",
			Message:  msg,
			Username: username,
		}
		response, err := json.Marshal(resMessage)
		if err != nil {
			log.Println(err)
			return
		}

		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			fmt.Printf("send message to client error: %s\n", err)
			continue
		}
	}
}
