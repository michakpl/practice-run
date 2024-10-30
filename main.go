package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Request struct {
	Action   string `json:"action"`
	Username string `json:"username"`
	Room     string `json:"room"`
	Message  string `json:"message"`
}

type Response struct {
	Action   string `json:"action"`
	Username string `json:"username"`
	Room     string `json:"room"`
	Message  string `json:"message"`
}

func main() {
	chat := &Chat{
		rooms: make(map[string]*Room),
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(chat, w, r)
	})
	err := http.ListenAndServe(":8999", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serveWs(chat *Chat, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
	}

	for {
		messageType, raw, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			continue
		}

		if messageType != websocket.TextMessage {
			log.Println("invalid message type")
			continue
		}

		var req Request
		if err := json.Unmarshal(raw, &req); err != nil {
			log.Println(err)
			continue
		}

		if req.Action == "createRoom" {
			err := chat.CreateRoom(req.Room)
			if err != nil {
				log.Println(err)
				continue
			}

			resMessage := Response{
				Action:  req.Action,
				Message: fmt.Sprintf("created room with name '%s'", req.Room),
				Room:    req.Room,
			}
			response, err := json.Marshal(resMessage)
			if err != nil {
				log.Println(err)
				continue
			}
			if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Println(err)
				continue
			}

			continue
		}

		if req.Action == "joinRoom" {
			_, err := chat.JoinRoom(client, req.Room)
			if err != nil {
				log.Println(err)
				continue
			}

			resMessage := Response{
				Action:  req.Action,
				Message: fmt.Sprintf("joined room with name '%s'", req.Room),
				Room:    req.Room,
			}
			response, err := json.Marshal(resMessage)
			if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Println(err)
				continue
			}

			continue
		}

		if req.Action == "leaveRoom" {
			err := chat.LeaveRoom(client, req.Room)
			if err != nil {
				log.Println(err)
				continue
			}

			resMessage := Response{
				Action:  req.Action,
				Message: fmt.Sprintf("left room with name '%s'", req.Room),
				Room:    req.Room,
			}
			response, err := json.Marshal(resMessage)
			if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Println(err)
				return
			}

			continue
		}

		if req.Action == "sendMessage" {
			room, err := chat.GetRoom(req.Room)
			if err != nil {
				log.Println(err)
				continue
			}

			room.SendMessage(req.Username, req.Message)

			continue
		}

		resMessage := Response{
			Action:  req.Action,
			Message: fmt.Sprintf("unexpected action '%s'", req.Action),
		}
		response, err := json.Marshal(resMessage)
		if err := client.WriteMessage(websocket.TextMessage, response); err != nil {
			log.Println(err)
			continue
		}
	}
}
