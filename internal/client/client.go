package client

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/n4vxn/Console-Chat/internal/ui"
)

var (
	WSAddr = "ws://localhost:8080/ws"
)

func WSClient() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	conn, _, err := websocket.DefaultDialer.Dial(WSAddr, nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket:", err)
		return
	}
	ui.Uitest()

	defer conn.Close()
	fmt.Println("Connected to WebSocket server")

	messageChannel := make(chan string)

	go handleRedisMessages(rdb, messageChannel)
	go handleWebSocketMessages(conn)
	go writeToWebSocket(conn, messageChannel)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		message := scanner.Text()

		messageChannel <- message
	}
}

func handleRedisMessages(rdb *redis.Client, messageChannel chan string) {
	ctx := context.Background()
	pubsub := rdb.Subscribe(ctx, "user-events")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		messageChannel <- msg.Payload
	}
}

func handleWebSocketMessages(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			return
		}
		fmt.Println(string(msg))
	}
}

func writeToWebSocket(conn *websocket.Conn, messageChannel chan string) {
	for message := range messageChannel {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error writing message to WebSocket:", err)
			return
		}
	}
}
