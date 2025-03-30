package websocket

import (
	"encoding/json"
	"github.com/gofiber/contrib/websocket"
	"log"
)

type miniClient map[string]map[string]*websocket.Conn // Modified type

type ClientObject struct {
	ChatID string
	UserID string
	Conn   *websocket.Conn
}

type BroadcastObject struct {
	MSG  string
	FROM ClientObject
}

var clients = make(miniClient) // Initialized as a nested map
var Register = make(chan ClientObject)
var Broadcast = make(chan BroadcastObject)
var Unregister = make(chan ClientObject)

func removeClient(org string, user string) {
	if conn, ok := clients[org][user]; ok { // Check if client exists
		delete(clients[org], user)
		conn.Close() // Close the connection before potentially removing the organization map
		if len(clients[org]) == 0 {
			delete(clients, org) // Remove empty organization map
		}
	}
}

func SocketHandler() {
	for {
		select {
		case client := <-Register:
			// Pre-initialize organization map if it doesn't exist
			if clients[client.ChatID] == nil {
				clients[client.ChatID] = make(map[string]*websocket.Conn)
			}
			clients[client.ChatID][client.UserID] = client.Conn
			log.Println("client registered:", client.ChatID, client.UserID)

		case message := <-Broadcast:
			for org, users := range clients {
				if org == message.FROM.ChatID {
					for user, conn := range users {
						if org != message.FROM.ChatID || user != message.FROM.UserID {
							msgPayload, err := json.Marshal(map[string]interface{}{
								"message": message.MSG,
								"from":    message.FROM,
							})
							if err != nil {
								log.Println("Error marshalling message:", err)
								continue
							}
							if err := conn.WriteMessage(websocket.TextMessage, msgPayload); err != nil {
								log.Println("write error:", err)
								removeClient(org, user) // Update client removal
								conn.WriteMessage(websocket.CloseMessage, []byte{})
								conn.Close()
							}
						}
					}
				}
			}

		case client := <-Unregister:
			removeClient(client.ChatID, client.UserID) // Update client removal
			log.Println("client Unregistered:", client.ChatID, client.UserID)
		}
	}
}
