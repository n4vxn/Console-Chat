package websocket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	"github.com/n4vxn/Console-Chat/internal/kafkaserver"
	"github.com/n4vxn/Console-Chat/internal/redis"
)

type ClientInfo struct {
	Username string
}

type WebSocketServer struct {
	clients       map[*websocket.Conn]ClientInfo // Track client info
	clientMutex   sync.Mutex                    // Protects the clients map
	kafkaProducer sarama.SyncProducer
	kafkaConsumer sarama.Consumer
}

func StartProducer(kafkaConfig *kafkaserver.KafkaConfig) (sarama.SyncProducer, error) {
	producer, err := sarama.NewSyncProducer([]string{kafkaConfig.Broker}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}
	return producer, nil
}

func StartConsumer(kafkaConfig *kafkaserver.KafkaConfig) (sarama.Consumer, error) {
	consumer, err := sarama.NewConsumer([]string{kafkaConfig.Broker}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}
	return consumer, nil
}

func NewWebSocketServer(kafkaProducer sarama.SyncProducer, kafkaConsumer sarama.Consumer) *WebSocketServer {
	return &WebSocketServer{
		clients:       make(map[*websocket.Conn]ClientInfo),
		kafkaProducer: kafkaProducer,
		kafkaConsumer: kafkaConsumer,
	}
}

func (ws *WebSocketServer) AddUsers(conn *websocket.Conn, username string) {
	ws.clientMutex.Lock()
	ws.clients[conn] = ClientInfo{Username: username}
	ws.clientMutex.Unlock()

	clientConnected := fmt.Sprintf("%s connected as %s", conn.RemoteAddr().String(), username)
	redis.PubUserEvent("user-events", clientConnected)
	log.Println(clientConnected)
}

func (ws *WebSocketServer) RemoveUsers(conn *websocket.Conn) {
	ws.clientMutex.Lock()
	clientInfo, exists := ws.clients[conn]
	if exists {
		delete(ws.clients, conn)
	}
	ws.clientMutex.Unlock()

	if exists {
		clientDisconnected := fmt.Sprintf("%s: disconnected", clientInfo.Username)
		redis.PubUserEvent("user-events", clientDisconnected)
		log.Println(clientDisconnected)
	}

	conn.Close()
}

func (ws *WebSocketServer) broadcastMessage(message string) {
	ws.clientMutex.Lock()
	defer ws.clientMutex.Unlock()

	for client := range ws.clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error sending message to client %v: %v", client.RemoteAddr(), err)
			client.Close()
			delete(ws.clients, client)
		}
	}
}

func (ws *WebSocketServer) HandleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "guest" // Default username if not provided
	}

	ws.AddUsers(conn, username)
	defer ws.RemoveUsers(conn)

	go func() {
		partitionConsumer, err := ws.kafkaConsumer.ConsumePartition("chat-messages", 0, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error consuming Kafka partition: %v", err)
			return
		}
		defer partitionConsumer.Close()

		for msg := range partitionConsumer.Messages() {
			message := fmt.Sprintf("%s: %s", username, string(msg.Value)) // Adjust to your actual Kafka message format
			ws.broadcastMessage(message)
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			break
		}

		kafkaMessage := &sarama.ProducerMessage{
			Topic: "chat-messages",
			Value: sarama.StringEncoder(msg),
		}

		_, _, err = ws.kafkaProducer.SendMessage(kafkaMessage)
		if err != nil {
			log.Printf("Error sending message to Kafka: %v", err)
		}
	}
}

