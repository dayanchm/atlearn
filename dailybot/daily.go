package main

import (
	"crypto/subtle"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

//go:embed sql/*/*.sql
var poetryFiles embed.FS

// Cache parsed data in memory, but keep no local publication state.
var dailyPoems = sync.OnceValues(func() ([]Poetry, error) {
	var poems []Poetry
	err := fs.WalkDir(poetryFiles, "sql", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := poetryFiles.ReadFile(path)
		if err != nil {
			return err
		}
		loaded, err := parsePoetrySQL(string(data))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		for _, poem := range loaded {
			text := poemPostText(poem)
			// Rune count is a conservative upper bound on grapheme count.
			if utf8.ValidString(text) && utf8.RuneCountInString(text) <= 300 && len(text) <= 3000 {
				poems = append(poems, poem)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(poems) == 0 {
		return nil, fmt.Errorf("no complete poems fit in one post")
	}
	return poems, nil
})

func poemPostText(poem Poetry) string {
	return poem.Title + "\n\n" + poem.Text + "\n\n— " + poem.Author
}

// A fixed UTC+5 zone avoids requiring an OS timezone database on Vercel.
var ashgabat = time.FixedZone("Asia/Ashgabat", 5*60*60)

func dailyDate(now time.Time) time.Time {
	local := now.In(ashgabat)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, ashgabat)
}

// A stable TID per local date prevents duplicate records across cron invocations.
// TIDs encode microseconds plus a 10-bit clock ID using sortable base32.
func dailyRecordKey(day time.Time) string {
	const alphabet = "234567abcdefghijklmnopqrstuvwxyz"
	value := uint64(day.UnixMicro()) << 10
	var key [13]byte
	for i := len(key) - 1; i >= 0; i-- {
		key[i] = alphabet[value&31]
		value >>= 5
	}
	return string(key[:])
}

func publishDailyPoem(now time.Time) (*PostResponse, error) {
	poems, err := dailyPoems()
	if err != nil {
		return nil, err
	}
	day := dailyDate(now)
	// Stable selection for a given date and dataset, rotating through short poems.
	index := (day.Unix() / 86400) % int64(len(poems))
	session, err := CreateSession(os.Getenv("BSKY_HANDLE"), os.Getenv("BSKY_APP_PASSWORD"))
	if err != nil {
		return nil, err
	}
	return createPostWithKey(session, poemPostText(poems[index]), dailyRecordKey(day))
}

func dailyHandler(publish func(time.Time) (*PostResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		secret := os.Getenv("CRON_SECRET")
		if strings.TrimSpace(secret) == "" {
			http.Error(w, "CRON_SECRET is not configured", http.StatusServiceUnavailable)
			return
		}
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+secret)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		result, err := publish(time.Now())
		if err != nil {
			log.Printf("daily poem failed: %v", err)
			http.Error(w, "daily post failed; check function logs before retrying", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
