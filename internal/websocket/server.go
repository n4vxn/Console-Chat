package websocket

import (
	"fmt"
	"log"
	"net/http"
	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	"github.com/n4vxn/Console-Chat/internal/kafkaserver"
)

type WebSocketServer struct {
	clients        map[*websocket.Conn]bool
	kafkaProducer  sarama.SyncProducer
	kafkaConsumer  sarama.Consumer
	broadcastChan  chan []byte
}

func NewWebSocketServer(kafkaProducer sarama.SyncProducer, kafkaConsumer sarama.Consumer) *WebSocketServer {
	return &WebSocketServer{
		clients:       make(map[*websocket.Conn]bool),
		kafkaProducer: kafkaProducer,
		kafkaConsumer: kafkaConsumer,
		broadcastChan: make(chan []byte),
	}
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
	defer conn.Close()

	ws.clients[conn] = true
	defer delete(ws.clients, conn)

	go func() {
		partitionConsumer, err := ws.kafkaConsumer.ConsumePartition("chat-messages", 0, sarama.OffsetNewest)
		if err != nil {
			log.Printf("Error consuming Kafka partition: %v", err)
			return
		}
		defer partitionConsumer.Close()

		for msg := range partitionConsumer.Messages() {
			message := fmt.Sprintf("Kafka Message: %s", string(msg.Value))
			for client := range ws.clients {
				err := client.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					log.Printf("Error sending message to client: %v", err)
					client.Close()
					delete(ws.clients, client)
				}
			}
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