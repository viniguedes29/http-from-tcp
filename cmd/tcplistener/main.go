package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		currentLine := ""

		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)

			if err != nil {
				break
			}

			buffer = buffer[:n]

			if i := bytes.IndexByte(buffer, '\n'); i != -1 {
				currentLine += string(buffer[:i])
				buffer = buffer[i+1:]
				ch <- currentLine
				currentLine = ""
			}

			currentLine += string(buffer)

		}

		if currentLine != "" {
			ch <- currentLine
		}
	}()

	return ch
}

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

		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}
	}
}
