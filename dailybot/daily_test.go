package main

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
	"unicode/utf8"
)

func TestDailyHandler(t *testing.T) {
	for _, tc := range []struct {
		name, method, secret, auth string
		fail                       bool
		status, calls              int
	}{
		{"missing secret", "GET", "", "", false, 503, 0},
		{"missing auth", "GET", "secret", "", false, 401, 0},
		{"wrong auth", "GET", "secret", "Bearer wrong", false, 401, 0},
		{"wrong method", "DELETE", "secret", "Bearer secret", false, 405, 0},
		{"cron", "GET", "secret", "Bearer secret", false, 200, 1},
		{"manual", "POST", "secret", "Bearer secret", false, 200, 1},
		{"upstream failure", "GET", "secret", "Bearer secret", true, 502, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CRON_SECRET", tc.secret)
			calls := 0
			handler := dailyHandler(func(time.Time) (*PostResponse, error) {
				calls++
				if tc.fail {
					return nil, fmt.Errorf("upstream error")
				}
				return &PostResponse{URI: "at://example", CID: "example"}, nil
			})
			req := httptest.NewRequest(tc.method, "/api/daily", nil)
			req.Header.Set("Authorization", tc.auth)
			response := httptest.NewRecorder()
			handler(response, req)
			if response.Code != tc.status || calls != tc.calls {
				t.Fatalf("status=%d calls=%d", response.Code, calls)
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Error("missing cache control")
			}
		})
	}
}

func TestEmbeddedDailyPoems(t *testing.T) {
	poems, err := scheduledPoems()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d poem excerpts fit a single post", len(poems))
	for _, p := range poems {
		text := poemPostText(p)
		if utf8.RuneCountInString(text) > 300 || len(text) > 3000 {
			t.Fatal("oversized poem selected")
		}
	}
}

func TestPostingRecordKey(t *testing.T) {
	before := time.Date(2026, 9, 7, 18, 59, 0, 0, time.UTC)
	after := before.Add(time.Minute)
	slot := postingSlot(before)
	if slot.Hour() != 18 || postingSlot(after).Hour() != 19 {
		t.Fatal("incorrect hour boundary")
	}
	key := postingRecordKey(slot)
	if len(key) != 13 || key != postingRecordKey(postingSlot(before.Add(-10*time.Minute))) || key == postingRecordKey(postingSlot(after)) {
		t.Fatal("unstable posting record key")
	}
	const alphabet = "234567abcdefghijklmnopqrstuvwxyz"
	var decoded uint64
	for _, c := range key {
		found := false
		for i, a := range alphabet {
			if c == a {
				decoded = decoded<<5 | uint64(i)
				found = true
				break
			}
		}
		if !found {
			t.Fatal("invalid TID alphabet")
		}
	}
	if decoded>>10 != uint64(slot.UnixMicro()) {
		t.Fatal("incorrect TID timestamp")
	}
}

func TestFourLineExcerpt(t *testing.T) {
	poem := Poetry{Title: " Başlyk ", Author: " Şair ", Text: "\r\n Birinji \r\n\nIkinji\n \nÜçünji\fDördünji\nBäşinji"}
	want := "Başlyk\n\nBirinji\nIkinji\nÜçünji\nDördünji\n\n— Şair"
	if got := poemPostText(poem); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	poem.Text = "Ýeke setir"
	if got := poemPostText(poem); got != "Başlyk\n\nÝeke setir\n\n— Şair" {
		t.Fatalf("short poem: %q", got)
	}
}

func TestRequestedPoemExcerpt(t *testing.T) {
	poems, err := LoadPoetry("sql/bahargul-mejidova/poetry.sql")
	if err != nil {
		t.Fatal(err)
	}
	want := "GEL SÖHBET EDELI\n\nGel, ikimiz bile söhbet edeli,\nBagt hakda, söýgi hakda söz açyp.\nGel, ikimiz bile söhbet edeli,\nYşk diýlen zat kalpdan gitmesin öçüp.\n\n— Bahargül Mejidowa"
	for _, poem := range poems {
		if poem.Title == "GEL SÖHBET EDELI" {
			if got := poemPostText(poem); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
			return
		}
	}
	t.Fatal("example poem not found")
}

func TestPostingIntervals(t *testing.T) {
	for _, minute := range []int{20, 40, 60} {
		boundary := time.Date(2026, 9, 7, 10, minute, 0, 0, time.UTC)
		previous := postingRecordKey(postingSlot(boundary.Add(-time.Second)))
		current := postingRecordKey(postingSlot(boundary))
		if previous == current {
			t.Fatalf("key did not change at %v", boundary)
		}
		if current != postingRecordKey(postingSlot(boundary.Add(19*time.Minute))) {
			t.Fatal("key changed within interval")
		}
	}
}
