package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/taham8875/http-from-tcp/internal/headers"
	"github.com/taham8875/http-from-tcp/internal/request"
	"github.com/taham8875/http-from-tcp/internal/response"
	"github.com/taham8875/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) {

		// handle httpbin proxy requests
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			handleHTTPBinProxy(w, req)
			return
		}

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
		case "/video":
			video, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				errorBody := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Failed to read video file</p>
  </body>
</html>`)
				newHeaders := headers.NewHeaders()
				newHeaders.SetOverride("Content-Type", "text/html")
				newHeaders.SetOverride("Content-Length", fmt.Sprintf("%d", len(errorBody)))
				newHeaders.SetOverride("Connection", "close")
				w.WriteHeaders(newHeaders)
				w.WriteBody(errorBody)
				return
			}

			newHeaders := headers.NewHeaders()
			newHeaders.SetOverride("Content-Type", "video/mp4")
			newHeaders.SetOverride("Content-Length", fmt.Sprintf("%d", len(video)))
			newHeaders.SetOverride("Connection", "close")

			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(newHeaders)
			w.WriteBody(video)

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

func handleHTTPBinProxy(w *response.Writer, req *request.Request) {
	// extraxt the path after /httpbin
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	if path == "" {
		path = "/"
	}

	// construct the base url
	httpbinURL := "https://httpbin.org" + path

	// Make the request to the httpbin server
	resp, err := http.Get(httpbinURL)
	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		errorBody := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Failed to proxy request to httpbin.org</p>
  </body>
</html>`)
		headers := headers.NewHeaders()
		headers.SetOverride("Content-Type", "text/plain")
		headers.SetOverride("Content-Length", fmt.Sprintf("%d", len(errorBody)))
		headers.SetOverride("Connection", "close")

		w.WriteHeaders(headers)
		w.WriteBody(errorBody)
		return
	}

	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOK)

	newHeaders := headers.NewHeaders()
	newHeaders.SetOverride("Content-Type", resp.Header.Get("Content-Type"))
	newHeaders.SetOverride("Transfer-Encoding", "chunked")
	newHeaders.SetOverride("Connection", "close")

	// remove the content-length header if it exists as we use chunked encoding
	newHeaders.Delete("Content-Length")

	w.WriteHeaders(newHeaders)

	var fullBody []byte
	buffer := make([]byte, 32)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// store the chunked data for hash calculation
			fullBody = append(fullBody, buffer[:n]...)
			_, writeErr := w.WriteChunkedBody(buffer[:n])
			if writeErr != nil {
				log.Println("Error writing chunked body:", writeErr)
				break
			}
		}
		if err != nil {
			break
		}
	}

	// calculate sha256 hash of the response body
	hash := sha256.Sum256(fullBody)
	hashHex := fmt.Sprintf("%x", hash)

	trailers := headers.NewHeaders() // trailers are headers btw
	trailers.SetOverride("X-Content-SHA256", hashHex)
	trailers.SetOverride("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

	w.WriteTrailers(trailers)
}
