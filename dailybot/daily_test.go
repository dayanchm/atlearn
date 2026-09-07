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
	poems, err := dailyPoems()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d complete poems fit a single post", len(poems))
	for _, p := range poems {
		text := poemPostText(p)
		if utf8.RuneCountInString(text) > 300 || len(text) > 3000 {
			t.Fatal("oversized poem selected")
		}
	}
}

func TestDailyRecordKey(t *testing.T) {
	before := time.Date(2026, 9, 7, 18, 59, 0, 0, time.UTC)
	after := before.Add(time.Minute)
	day := dailyDate(before)
	if day.Day() != 7 || dailyDate(after).Day() != 8 {
		t.Fatal("incorrect Ashgabat date boundary")
	}
	key := dailyRecordKey(day)
	if len(key) != 13 || key != dailyRecordKey(dailyDate(before.Add(-time.Hour))) || key == dailyRecordKey(dailyDate(after)) {
		t.Fatal("unstable daily record key")
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
	if decoded>>10 != uint64(day.UnixMicro()) {
		t.Fatal("incorrect TID timestamp")
	}
}
