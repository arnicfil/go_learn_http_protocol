package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Printf("Error while resolving udp address: %v", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Printf("Error while dialing udp: %v", err)
		return
	}
	defer conn.Close()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">")
		line, err := reader.ReadString(10)
		if err != nil {
			fmt.Printf("Error reading line from stdin: %v", err)
			return
		}
		conn.Write([]byte(line))
	}
}
