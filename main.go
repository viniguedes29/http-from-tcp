package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
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
	file, err := os.Open("messages.txt")
	if err != nil {
		return
	}

	for line := range getLinesChannel(file) {
		fmt.Printf("read: %s\n", line)
	}
}
