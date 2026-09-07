package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestLocalMinuteSelection(t *testing.T) {
	now := time.Date(2026, 9, 7, 10, 21, 10, 0, time.UTC)
	slot, index := poemSlotAndIndex(now, time.Minute, 765)
	same, sameIndex := poemSlotAndIndex(now.Add(30*time.Second), time.Minute, 765)
	next, nextIndex := poemSlotAndIndex(now.Add(time.Minute), time.Minute, 765)
	if slot != same || index != sameIndex {
		t.Fatal("selection changed within minute")
	}
	if postingRecordKey(slot) == postingRecordKey(next) || nextIndex != (index+1)%765 {
		t.Fatal("minute did not advance key and poem")
	}
	remote, _ := poemSlotAndIndex(now, 20*time.Minute, 765)
	if remote != postingSlot(now) {
		t.Fatal("HTTP posting interval changed")
	}
}

func TestLocalScheduleContinuesAfterError(t *testing.T) {
	ticks := make(chan time.Time, 2)
	first := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	ticks <- first
	ticks <- first.Add(time.Minute)
	close(ticks)
	calls := 0
	runLocalSchedule(context.Background(), ticks, func(now time.Time) (*PostResponse, error) {
		calls++
		if now != first.Add(time.Duration(calls-1)*time.Minute) {
			t.Fatal("incorrect tick time")
		}
		if calls == 1 {
			return nil, fmt.Errorf("temporary failure")
		}
		return &PostResponse{URI: "at://example"}, nil
	})
	if calls != 2 {
		t.Fatalf("got %d calls", calls)
	}
}

func TestLocalScheduleCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ticks := make(chan time.Time, 1)
	ticks <- time.Now()
	runLocalSchedule(ctx, ticks, func(time.Time) (*PostResponse, error) { t.Fatal("published after cancellation"); return nil, nil })
}

func TestLocalSourceCredit(t *testing.T) {
	poem := Poetry{Title: "Başlyk", Text: "Bir\nIki\nÜç\nDört\nBäş", Author: "Şair"}
	want := "Başlyk\n\nBir\nIki\nÜç\nDört\n\n— Şair\n\nÇeşme: github.com/turkmenos/tm-data"
	if got := localPoemPostText(poem); got != want {
		t.Fatalf("got %q", got)
	}
	poems, err := localPoems()
	if err != nil {
		t.Fatal(err)
	}
	for _, poem := range poems {
		text := localPoemPostText(poem)
		if len([]rune(text)) > 300 || len(text) > 3000 {
			t.Fatal("source credit exceeds post limit")
		}
	}
}

func TestAuthorRotation(t *testing.T) {
	poems := []Poetry{
		{Author: "Kerim", Title: "K1"}, {Author: "Kerim", Title: "K2"}, {Author: "Kerim", Title: "K3"},
		{Author: "Bahargül", Title: "B1"}, {Author: "Magrupy", Title: "M1"}, {Author: "Magrupy", Title: "M2"},
	}
	want := []string{"K1", "B1", "M1", "K2", "B1", "M2", "K3", "B1", "M1", "K1", "B1", "M2"}
	for i, title := range want {
		now := time.Unix(int64(i)*60, 0)
		slot, poem, err := selectPoemByAuthor(now, time.Minute, poems)
		if err != nil || poem.Title != title {
			t.Fatalf("turn %d: got %+v, err=%v; want %s", i, poem, err, title)
		}
		sameSlot, same, err := selectPoemByAuthor(now.Add(59*time.Second), time.Minute, poems)
		if err != nil || slot != sameSlot || poem != same {
			t.Fatal("selection changed within minute")
		}
	}
}

func TestLocalDatasetRotatesAllAuthors(t *testing.T) {
	poems, err := localPoems()
	if err != nil {
		t.Fatal(err)
	}
	authors := map[string]bool{}
	for _, poem := range poems {
		authors[poem.Author] = true
	}
	if len(authors) < 2 {
		t.Fatal("not enough authors to rotate")
	}
	seen := map[string]bool{}
	start := time.Date(2026, 9, 7, 10, 0, 0, 0, time.UTC)
	for i := 0; i < len(authors); i++ {
		_, poem, err := selectPoemByAuthor(start.Add(time.Duration(i)*time.Minute), time.Minute, poems)
		if err != nil {
			t.Fatal(err)
		}
		if seen[poem.Author] {
			t.Fatalf("author repeated before cycle finished: %s", poem.Author)
		}
		seen[poem.Author] = true
	}
	t.Logf("Verified rotation across %d eligible authors", len(seen))
}
