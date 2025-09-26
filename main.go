package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		return
	}

	defer file.Close()

	buffer := make([]byte, 8)

	for {
		n, err := file.Read(buffer)
		if err != nil {
			return
		}

		fmt.Printf("read: %s\n", string(buffer[:n]))
	}
}
