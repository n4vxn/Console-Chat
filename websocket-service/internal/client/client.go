package client

import (
	"bufio"
	"fmt"
	"os"

	"github.com/gorilla/websocket"
)

var (
	WSAddr = "ws://localhost:8080/ws"
)

func WSClient() {
	conn, _, err := websocket.DefaultDialer.Dial(WSAddr, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)

	for {

		scanner.Scan()

		message := scanner.Text()
		
		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

}
