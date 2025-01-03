package main

import (
	"log"
	"net/http"

	"github.com/n4vxn/Console-Chat/internal/db"
	"github.com/n4vxn/Console-Chat/internal/kafkaserver"
	"github.com/n4vxn/Console-Chat/internal/redis"
	"github.com/n4vxn/Console-Chat/internal/websocket"
)

func main() {
	redis.InitRedis()
	kafkaConfig := &kafkaserver.KafkaConfig{
		Broker: "localhost:9092",
		Topic:  "chat-messages",
	}

	producer, err := websocket.StartProducer(kafkaConfig)
	if err != nil {
		log.Fatalf("Error starting Kafka producer: %v", err)
	}
	defer producer.Close()

	consumer, err := websocket.StartConsumer(kafkaConfig)
	if err != nil {
		log.Fatalf("Error starting Kafka consumer: %v", err)
	}
	defer consumer.Close()


	err = db.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	wsServer := websocket.NewWebSocketServer(producer, consumer)

	http.HandleFunc("/ws", wsServer.HandleWebSocketConnections)

	log.Println("Starting WebSocket server on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}

}
