package main

import (
	"socket-rsudlampung/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Mendefinisikan route GET untuk "/ws" yang akan memanggil fungsi `WsEndpoint` di package `handlers`
	router.GET("/ws", handlers.WsEndpoint)

	// Menjalankan fungsi `ListenToWsChannel` di dalam goroutine untuk menangani WebSocket secara asynchronous
	go handlers.ListenToWsChannel()

	// Menjalankan server pada port 8181
	router.Run(":8181")
}
