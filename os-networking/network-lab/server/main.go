package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	fmt.Println("Waiting Client ...")

	conn, err := listener.Accept()

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Locale:", conn.LocalAddr())
	fmt.Println("Remote:", conn.RemoteAddr())
	fmt.Println("Connected to Client...", conn.RemoteAddr())

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			fmt.Println("Client closed connected")
			return
		}

		if err != nil {
			fmt.Println("Read error:", err)
			return
		}
		fmt.Printf("Read byte: %d\n", n)
		fmt.Printf("Data: %q\n", string(buf[:n]))

	}

}
