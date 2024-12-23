package websocket

import (
	"bufio"
	"fmt"
	"os"
	"log"

	"github.com/gorilla/websocket"
)

var (
	WSAddr = "ws://localhost:8080/ws"
)

func WSClient() {
	conn, _, err := websocket.DefaultDialer.Dial(WSAddr, nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to WebSocket server")

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error reading message:", err)
				return
			}
			fmt.Printf("Received from server: %s\n", msg)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()

		message := scanner.Text()
		
		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}
