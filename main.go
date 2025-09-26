package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		return
	}

	defer file.Close()

	currentLine := ""

	for {
		buffer := make([]byte, 8)
		n, err := file.Read(buffer)

		if err != nil {
			break
		}

		buffer = buffer[:n]

		if i := bytes.IndexByte(buffer, '\n'); i != -1 {
			currentLine += string(buffer[:i])
			buffer = buffer[i+1:]
			fmt.Printf("read: %s\n", currentLine)
			currentLine = ""
		}

		currentLine += string(buffer)

	}

	if len(currentLine) != 0 {
		fmt.Printf("read: %s\n", currentLine)
	}
}
