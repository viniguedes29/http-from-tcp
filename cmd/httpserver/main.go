package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taham8875/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	s, err := server.Serve(port)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer s.Close()
	log.Printf("Server listening on port %d", port)

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)
	<-sigChannel
	log.Println("Shutting down server...")
}
