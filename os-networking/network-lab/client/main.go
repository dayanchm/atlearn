package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected to server")
	fmt.Println("Locale:", conn.LocalAddr())
	fmt.Println("Remote:", conn.RemoteAddr())

	data := []byte("hello from client")
	n, err := conn.Write(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Send %d bytes/n", n)
}
