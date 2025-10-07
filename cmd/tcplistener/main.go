package main

import (
	"fmt"
	"net"

	"github.com/taham8875/http-from-tcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		r, err := request.RequestFromReader(conn)

		if err != nil {
			fmt.Printf("Error reading request: %v\n", err)
			conn.Close()
			continue
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for k, v := range r.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
	}
}
