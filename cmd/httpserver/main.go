package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taham8875/http-from-tcp/internal/request"
	"github.com/taham8875/http-from-tcp/internal/response"
	"github.com/taham8875/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				Code:    response.StatusBadRequest,
				Message: "Your problem is not my problem\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				Code:    response.StatusInternalServerError,
				Message: "Woopsie, my bad\n",
			}
		default:
			_, _ = io.Copy(w, bytes.NewBufferString("All good, frfr\n"))
			return nil
		}
	}

	s, err := server.Serve(port, handler)
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
