package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"
)

func runLocalBot() error {
	if strings.TrimSpace(os.Getenv("BSKY_HANDLE")) == "" || strings.TrimSpace(os.Getenv("BSKY_APP_PASSWORD")) == "" {
		return fmt.Errorf("set BSKY_HANDLE and BSKY_APP_PASSWORD in .env")
	}
	poems, err := localPoems()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	log.Print("Local bot started. First post in one minute, then every minute. Press Ctrl+C to stop.")
	runLocalSchedule(ctx, ticker.C, func(now time.Time) (*PostResponse, error) {
		return publishPoemSelection(now, time.Minute, poems, localPoemPostText)
	})
	log.Print("Local bot stopped.")
	return nil
}

// Publish synchronously: slow requests cannot overlap with another local post.
func runLocalSchedule(ctx context.Context, ticks <-chan time.Time, publish func(time.Time) (*PostResponse, error)) {
	for {
		select {
		case <-ctx.Done():
			return
		case now, ok := <-ticks:
			if !ok || ctx.Err() != nil {
				return
			}
			result, err := publish(now)
			if err != nil {
				log.Printf("Local post failed: %v", err)
				continue
			}
			log.Printf("Published: %s", result.URI)
		}
	}
}

func localPoemPostText(poem Poetry) string {
	return poemPostText(poem) + "\n\nÇeşme: github.com/turkmenos/tm-data"
}

var localPoems = sync.OnceValues(func() ([]Poetry, error) {
	poems, err := scheduledPoems()
	if err != nil {
		return nil, err
	}
	var eligible []Poetry
	for _, poem := range poems {
		text := localPoemPostText(poem)
		if utf8.RuneCountInString(text) <= 300 && len(text) <= 3000 {
			eligible = append(eligible, poem)
		}
	}
	if len(eligible) == 0 {
		return nil, fmt.Errorf("no excerpts fit with the source credit")
	}
	return eligible, nil
})
