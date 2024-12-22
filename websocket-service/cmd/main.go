package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/n4vxn/Console-Chat/websocket-service/internal/client"
	"github.com/n4vxn/Console-Chat/websocket-service/internal/server"
)

var (
	WSAddr = "ws://localhost:8080/ws"
)

func main() {
	wsServer := server.NewWebSocketServer()
	go wsServer.BroadcastMessages()
	r := gin.Default()

	r.GET("/ws", wsServer.HandleWS)

	log.Printf("server started at %s", WSAddr)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}

	client.WSClient()
}
