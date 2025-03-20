package websocket

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	"log"
)

type miniClient map[string]*websocket.Conn // Modified type

type ClientObject struct {
	UserID string
	Conn   *websocket.Conn
}

type BroadcastObject struct {
	MSG       map[string]interface{}
	Recipient string
}

var clients = make(miniClient) // Initialized as a nested map
var Register = make(chan ClientObject)
var Broadcast = make(chan BroadcastObject)
var Unregister = make(chan ClientObject)

func removeClient(user string) {
	if conn, ok := clients[user]; ok { // Check if client exists
		delete(clients, user)
		conn.Close()
	}
}

func SocketHandler() {
	for {
		select {
		case client := <-Register:
			clients[client.UserID] = client.Conn
			log.Println("client registered:", client.UserID)

		case message := <-Broadcast:
			for user, conn := range clients {
				if user == message.Recipient {
					msgPayload, err := json.Marshal(map[string]interface{}{
						"Data": message.MSG,
					})
					if err != nil {
						log.Println("Error marshalling message:", err)
						continue
					}
					// Send the message
					if err := conn.WriteMessage(websocket.TextMessage, msgPayload); err != nil {
						log.Println("write error:", err)
						removeClient(user)
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						conn.Close()
					}

				}

			}

		case client := <-Unregister:
			log.Println("client Unregistered:", client.Conn)
		}
	}
}
