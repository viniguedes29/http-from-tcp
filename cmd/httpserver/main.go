package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taham8875/http-from-tcp/internal/headers"
	"github.com/taham8875/http-from-tcp/internal/request"
	"github.com/taham8875/http-from-tcp/internal/response"
	"github.com/taham8875/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			w.WriteStatusLine(response.StatusBadRequest)
			body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
			headers := headers.NewHeaders()
			headers.SetOverride("Content-Type", "text/html")
			headers.SetOverride("Content-Length", fmt.Sprintf("%d", len(body)))
			headers.SetOverride("Connection", "close")

			w.WriteHeaders(headers)
			w.WriteBody(body)
		case "/myproblem":
			w.WriteStatusLine(response.StatusInternalServerError)

			body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
			headers := headers.NewHeaders()
			headers.SetOverride("Content-Type", "text/html")
			headers.SetOverride("Content-Length", fmt.Sprintf("%d", len(body)))
			headers.SetOverride("Connection", "close")

			w.WriteHeaders(headers)
			w.WriteBody(body)
		default:
			w.WriteStatusLine(response.StatusOK)

			body := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

			headers := headers.NewHeaders()
			headers.SetOverride("Content-Type", "text/html")
			headers.SetOverride("Content-Length", fmt.Sprintf("%d", len(body)))
			headers.SetOverride("Connection", "close")

			w.WriteHeaders(headers)
			w.WriteBody(body)
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
