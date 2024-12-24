package websocket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	"github.com/n4vxn/Console-Chat/internal/kafkaserver"
)

type ClientInfo struct {
	Username string
}

type WebSocketServer struct {
	clients       map[*websocket.Conn]bool
	clientMutex   sync.Mutex // Protects the clients map
	clientInfo    ClientInfo
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
		clients:       make(map[*websocket.Conn]bool),
		kafkaProducer: kafkaProducer,
		kafkaConsumer: kafkaConsumer,
	}
}

func (ws *WebSocketServer) AddUsers(conn *websocket.Conn) {
	ws.clientMutex.Lock()
	ws.clients[conn] = true
	ws.clientMutex.Unlock()
	// defer func() {
	// 	ws.clientMutex.Lock()
	// 	delete(ws.clients, conn)
	// 	ws.clientMutex.Unlock()
	// 	conn.Close()
	// 	log.Printf("%s disconnected", conn.RemoteAddr().String())
	// }()

	log.Printf("%s connected", conn.RemoteAddr().String())
}

func (ws *WebSocketServer) RemoveUsers(conn *websocket.Conn) {
	defer func() {
		ws.clientMutex.Lock()
		delete(ws.clients, conn)
		ws.clientMutex.Unlock()
		conn.Close()
		log.Printf("%s disconnected", conn.RemoteAddr().String())
	}()
}

func (ws *WebSocketServer) broadcastMessage(message string) {
	ws.clientMutex.Lock()
	defer ws.clientMutex.Unlock()

	for client := range ws.clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
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

	ws.AddUsers(conn)
	defer ws.RemoveUsers(conn)

	go func() {
		partitionConsumer, err := ws.kafkaConsumer.ConsumePartition("chat-messages", 0, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error consuming Kafka partition: %v", err)
			return
		}
		defer partitionConsumer.Close()
		ws.clientInfo.Username = "naveen"
		for msg := range partitionConsumer.Messages() {
			message := fmt.Sprintf("%s: %s", ws.clientInfo.Username, string(msg.Value))
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
