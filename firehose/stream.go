package main

import (
	"context"
	"fmt"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/events"
	"github.com/bluesky-social/indigo/events/schedulers/sequential"
	"github.com/gorilla/websocket"
)

const firehoseURL = "wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos"

func RunStream(ctx context.Context) error {
	fmt.Println("Connecting to firehose...")

	conn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		firehoseURL,
		nil,
	)
	if err != nil {
		return fmt.Errorf("connect firehose: %w", err)
	}
	defer conn.Close()

	fmt.Println("Connected:", firehoseURL)

	callbacks := &events.RepoStreamCallbacks{
		RepoCommit: func(event *comatproto.SyncSubscribeRepos_Commit) error {
			fmt.Println("COMMIT")
			fmt.Println("repo:", event.Repo)
			fmt.Println("seq:", event.Seq)
			fmt.Println("rev:", event.Rev)
			fmt.Println("ops:", len(event.Ops))

			for _, op := range event.Ops {
				recordType := RecordTypeFromPath(op.Path)
				fmt.Println("action:", op.Action)
				fmt.Println("path:", op.Path)
				fmt.Println("type:", recordType)
			}
			fmt.Println("---")
			return nil
		},
	}

	scheduler := sequential.NewScheduler("firehose", callbacks.EventHandler)

	return events.HandleRepoStream(ctx, conn, scheduler, nil)
}
