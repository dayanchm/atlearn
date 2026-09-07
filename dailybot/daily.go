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
var scheduledPoems = sync.OnceValues(func() ([]Poetry, error) {
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
		return nil, fmt.Errorf("no four-line excerpts fit in one post")
	}
	return poems, nil
})

func poemPostText(poem Poetry) string {
	lines := make([]string, 0, 4)
	for _, line := range strings.FieldsFunc(poem.Text, func(r rune) bool { return r == '\n' || r == '\r' || r == '\f' }) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) == 4 {
			break
		}
	}
	return strings.TrimSpace(poem.Title) + "\n\n" + strings.Join(lines, "\n") + "\n\n— " + strings.TrimSpace(poem.Author)
}

// Use the same 60-minute slot for selection and duplicate prevention.
func postingSlot(now time.Time) time.Time {
	return now.UTC().Truncate(60 * time.Minute)
}

// A stable TID per posting slot prevents duplicate records across cron invocations.
// TIDs encode microseconds plus a 10-bit clock ID using sortable base32.
func postingRecordKey(slot time.Time) string {
	const alphabet = "234567abcdefghijklmnopqrstuvwxyz"
	value := uint64(slot.UnixMicro()) << 10
	var key [13]byte
	for i := len(key) - 1; i >= 0; i-- {
		key[i] = alphabet[value&31]
		value >>= 5
	}
	return string(key[:])
}

func publishScheduledPoem(now time.Time) (*PostResponse, error) {
	return publishPoemAtInterval(now, 60*time.Minute)
}

func publishPoemAtInterval(now time.Time, interval time.Duration) (*PostResponse, error) {
	poems, err := scheduledPoems()
	if err != nil {
		return nil, err
	}
	return publishPoemSelection(now, interval, poems, poemPostText)
}

func publishPoemSelection(now time.Time, interval time.Duration, poems []Poetry, format func(Poetry) string) (*PostResponse, error) {
	slot, poem, err := selectPoemByAuthor(now, interval, poems)
	if err != nil {
		return nil, err
	}
	session, err := CreateSession(os.Getenv("BSKY_HANDLE"), os.Getenv("BSKY_APP_PASSWORD"))
	if err != nil {
		return nil, err
	}
	return createPostWithKey(session, format(poem), postingRecordKey(slot))
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
			log.Printf("scheduled poem failed: %v", err)
			http.Error(w, "scheduled post failed; check function logs before retrying", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

// Selection and record keys share the interval so local minute posts advance together.
func poemSlotAndIndex(now time.Time, interval time.Duration, count int) (time.Time, int) {
	slot := now.UTC().Truncate(interval)
	return slot, int((slot.UnixNano() / int64(interval)) % int64(count))
}

// Give each author one turn per cycle, regardless of their number of poems.
// Stable ordering makes retries and restarts within a slot select the same poem.
func selectPoemByAuthor(now time.Time, interval time.Duration, poems []Poetry) (time.Time, Poetry, error) {
	if interval <= 0 || len(poems) == 0 {
		return time.Time{}, Poetry{}, fmt.Errorf("a positive interval and eligible poems are required")
	}
	authors := make([]string, 0)
	grouped := make(map[string][]Poetry)
	for _, poem := range poems {
		if _, exists := grouped[poem.Author]; !exists {
			authors = append(authors, poem.Author)
		}
		grouped[poem.Author] = append(grouped[poem.Author], poem)
	}
	slot := now.UTC().Truncate(interval)
	turn := slot.UnixNano() / int64(interval)
	if turn < 0 {
		return time.Time{}, Poetry{}, fmt.Errorf("posting time must not precede Unix epoch")
	}
	author := authors[turn%int64(len(authors))]
	choices := grouped[author]
	index := (turn / int64(len(authors))) % int64(len(choices))
	return slot, choices[index], nil
}
