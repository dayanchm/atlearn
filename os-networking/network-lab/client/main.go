package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func main() {
	/** Conn() **/
	Dialer()
}

func Dialer() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	dialer := net.Dialer{}

	conn, err := dialer.DialContext(

		ctx,
		"tcp",
		"localhost:8080",
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("1. Connected")

	_, err = conn.Write([]byte("hello"))
	if err != nil {
		panic(err)
	}
	fmt.Println("2. Connected")

	time.Sleep(5 * time.Second)

	fmt.Println("3. Closing connection")
	err = conn.Close()
	if err != nil {
		fmt.Println("close error", err)
		return
	}
	fmt.Println("4. Connection closed")
}

/** net.Conn lifecycle **/

func Conn() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	fmt.Println("1. Connected")

	_, err = conn.Write([]byte("hello"))
	if err != nil {
		panic(err)
	}
	fmt.Println("2. Connected")

	time.Sleep(5 * time.Second)

	fmt.Println("3. Closing connection")
	err = conn.Close()
	if err != nil {
		fmt.Println("close error", err)
		return
	}
	fmt.Println("4. Connection closed")

}
