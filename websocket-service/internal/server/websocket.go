package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketServer struct {
	clients   map[*Client]bool
	broadcast chan BroadcastMessage
	mux       sync.Mutex
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[*Client]bool),
		broadcast: make(chan BroadcastMessage),
	}
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	username string
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte),
	}
}

func (ws *WebSocketServer) AddClient(client *Client) {
	ws.mux.Lock()
	defer ws.mux.Unlock()
	client.username = "naveen"
	ws.clients[client] = true

	log.Printf("%s connected!", client.conn.RemoteAddr().String())
}

func (ws *WebSocketServer) RemoveClient(client *Client) {
	ws.mux.Lock()
	defer ws.mux.Unlock()
	if _, ok := ws.clients[client]; ok {
		delete(ws.clients, client)
		close(client.send)
		log.Printf("Client %s disconnected!", client.conn.RemoteAddr().String())
	}
	log.Printf("%s connected!", client.conn.RemoteAddr().String())
}

type BroadcastMessage struct {
	client *Client
	msg    []byte
}

func (ws *WebSocketServer) HandleClientMessages(client *Client) {
	defer ws.RemoveClient(client)

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %v", client.conn.RemoteAddr().String(), err)
			break
		}
		ws.broadcast <- BroadcastMessage{client: client, msg: msg}
	}
}

func (s *WebSocketServer) BroadcastMessages() {
	for {
		broadcastMsg := <-s.broadcast
		s.mux.Lock()
		for client := range s.clients {
			if client != broadcastMsg.client {
				select {
				case client.send <- broadcastMsg.msg:
				default:
					s.RemoveClient(client)
				}

			}
		}
		s.mux.Unlock()
	}
}

func (ws *WebSocketServer) SendMessagesToClient(client *Client) {
	defer client.conn.Close()

	for msg := range client.send {
		message := fmt.Sprintf("%s: %s", client.username, string(msg))
		err := client.conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error sending message to client %s: %v", client.conn.RemoteAddr().String(), err)
			break
		}
	}
}

func (ws *WebSocketServer) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn)
	ws.AddClient(client)

	go ws.HandleClientMessages(client)
	go ws.SendMessagesToClient(client)
}
