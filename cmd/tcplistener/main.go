package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		fmt.Printf("Error while creating listener: %v", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error while accepting from listener: %v", err)
			return
		}
		fmt.Println("Connection has been accepted")

		ch := getLinesChannel(conn)
		for line := range ch {
			fmt.Println("read: " + line)
		}

		fmt.Println("Connection has been closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	stChan := make(chan string)
	go routine(stChan, f)

	return stChan
}

func routine(ch chan string, f io.ReadCloser) {
	defer f.Close()
	chunkSize := 8
	buffer := make([]byte, chunkSize)
	var line string

	for {
		var chunk string
		n, err := f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				ch <- line
				break
			}

			log.Fatalf("Error while reading: %v", err)
		}

		chunk = string(buffer[:n])
		parts := strings.Split(chunk, "\n")

		if len(parts) > 1 {
			for i, part := range parts {
				if i == 0 {
					line = line + part
					ch <- line
				} else if i == len(parts)-1 {
					line = part
				} else {
					ch <- part
				}
			}
		} else {
			line = line + chunk
		}
	}

	close(ch)
}
