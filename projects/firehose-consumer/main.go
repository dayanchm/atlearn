package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coder/websocket"
)

const endpoint = "wss://jetstream1.us-east.bsky.network/subscribe"

type Event struct {
	DID    string          `json:"did"`
	TimeUs int64           `json:"time_us"`
	Kind   string          `json:"kind"`
	Commit *Commit         `json:"commit,omitempty"`
	Raw    json.RawMessage `json:"-"`
}

type Commit struct {
	Rev        string          `json:"rev"`
	Operation  string          `json:"operation"`
	Collection string          `json:"collection"`
	RKey       string          `json:"rkey"`
	Record     json.RawMessage `json:"record"`
	CID        string          `json:"cid"`
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	fmt.Println("Connecting to Jetstream ...")

	conn, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close(websocket.StatusNormalClosure, "shutdown")

	fmt.Println("Connected")
	fmt.Println()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				fmt.Println("Shutting down ...")
				return
			}
			log.Fatal(err)
		}
		var event Event
		if err := json.Unmarshal(data, &event); err != nil {
			log.Println("decode:", err)
			continue
		}
		printEvent(event)

	}
}

func printEvent(event Event) {
	if event.Commit == nil {
		return
	}

	fmt.Printf(
		"%s  %-6s  %s\n",
		event.DID,
		event.Commit.Operation,
		event.Commit.Collection,
	)
}
