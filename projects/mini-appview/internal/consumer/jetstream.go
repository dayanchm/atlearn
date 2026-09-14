package consumer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder/websocket"

	"github.com/dayanchm/mini-appview/internal/store"
)

type Event struct {
	DID    string  `json:"did"`
	Commit *Commit `json:"commit,omitempty"`
}

type Commit struct {
	Operation  string          `json:"operation"`
	Collection string          `json:"collection"`
	RKey       string          `json:"rkey"`
	CID        string          `json:"cid"`
	Record     json.RawMessage `json:"record"`
}

type PostRecord struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

func Run(ctx context.Context, db *sql.DB) error {

	const endpoint = "wss://jetstream1.us-east.bsky.network/subscribe"
	conn, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "shutdown")
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		var event Event
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}
		if event.Commit == nil {
			continue
		}
		if event.Commit.Collection != "app.bsky.feed.post" {
			continue
		}
		if event.Commit.Operation != "create" {
			continue
		}
		if err := handlePost(db, event); err != nil {
			fmt.Println("store post:", err)
		}
	}

}
func handlePost(db *sql.DB, event Event) error {

	var record PostRecord
	if err := json.Unmarshal(event.Commit.Record, &record); err != nil {
		return err
	}
	createdAt, err := time.Parse(time.RFC3339Nano, record.CreatedAt)
	if err != nil {
		return err
	}
	uri := fmt.Sprintf(
		"at://%s/%s/%s",
		event.DID,
		event.Commit.Collection,
		event.Commit.RKey,
	)
	post := store.Post{
		URI:       uri,
		CID:       event.Commit.CID,
		AuthorDID: event.DID,
		Text:      record.Text,
		CreatedAt: createdAt,
	}
	return store.SavePost(db, post)

}
