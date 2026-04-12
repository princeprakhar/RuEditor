package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rueditor/backend/internal/api"
	"github.com/rueditor/backend/internal/services"
	"github.com/rueditor/backend/internal/websocket"
)

func main() {
	latexService := services.NewLatexService()
	wsManager := websocket.NewManager()

	go wsManager.Run()

	wsServer := websocket.NewWSServer(wsManager, "8081")
	go func() {
		if err := wsServer.StartServer(); err != nil {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	apiServer := api.NewServer(latexService, wsManager)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := apiServer.Run(); err != nil {
			log.Fatal("Failed to start API server:", err)
		}
	}()

	log.Println("Server started on :8080 (HTTP) and :8081 (WebSocket)")
	<-quit
	log.Println("Shutting down server...")
}
