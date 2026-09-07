package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSessionAndPoetryPost(t *testing.T) {
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("invalid HTTP request")
		}
		switch r.URL.Path {
		case "/xrpc/com.atproto.server.createSession":
			var input SessionRequest
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Error(err)
			}
			if input.Identifier != "poet.example" || input.Password != "app-password" {
				t.Error("incorrect login body")
			}
			fmt.Fprint(w, `{"did":"did:plc:test","handle":"poet.example","accessJwt":"test-token","refreshJwt":"refresh-token"}`)
		case "/xrpc/com.atproto.repo.createRecord":
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Error("missing bearer token")
			}
			var input struct {
				Repo       string
				Collection string
				Record     struct {
					Type      string `json:"$type"`
					Text      string
					CreatedAt string
				}
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Error(err)
			}
			if input.Repo != "did:plc:test" || input.Collection != "app.bsky.feed.post" || input.Record.Type != "app.bsky.feed.post" || input.Record.Text != "Başlyk\n\nGoşgy\n\n— Şair" {
				t.Errorf("invalid post: %+v", input)
			}
			if _, err := time.Parse(time.RFC3339Nano, input.Record.CreatedAt); err != nil {
				t.Error(err)
			}
			fmt.Fprint(w, `{"uri":"at://did:plc:test/app.bsky.feed.post/123","cid":"test-cid"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	old := postHTTPClient
	postHTTPClient = server.Client()
	t.Cleanup(func() { postHTTPClient = old })
	t.Setenv("BSKY_PDS_URL", server.URL)
	session, err := CreateSession("poet.example", "app-password")
	if err != nil {
		t.Fatal(err)
	}
	result, err := PostPoetry(session, Poetry{Title: "Başlyk", Text: "Goşgy", Author: "Şair"})
	if err != nil {
		t.Fatal(err)
	}
	if result.CID != "test-cid" || calls != 2 {
		t.Fatalf("result=%+v calls=%d", result, calls)
	}
	for _, text := range []string{"", "  ", strings.Repeat("a", 3001), string([]byte{0xff})} {
		if _, err := CreatePost(session, text); err == nil {
			t.Error("invalid text accepted")
		}
	}
	if calls != 2 {
		t.Fatal("invalid text reached server")
	}
}

func TestPostAPIErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"unauthorized", 401, `{"error":"AuthenticationRequired","message":"secret-password"}`},
		{"rate limit", 429, `{"error":"RateLimitExceeded"}`},
		{"invalid JSON", 200, `{`},
		{"missing fields", 200, `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			_, err := CreatePost(&Session{AccessJWT: "token", DID: "did:plc:test", service: server.URL}, "Goşgy")
			if err == nil {
				t.Fatal("expected error")
			}
			if strings.Contains(err.Error(), "secret-password") {
				t.Fatal("server message leaked")
			}
			if calls != 1 {
				t.Fatal("unexpected retry")
			}
		})
	}
	if _, err := CreatePost(nil, "Goşgy"); err == nil {
		t.Fatal("nil session accepted")
	}
	if _, err := CreateSession("", ""); err == nil {
		t.Fatal("empty credentials accepted")
	}
	t.Setenv("BSKY_PDS_URL", "http://example.com")
	if _, err := CreateSession("poet", "password"); err == nil {
		t.Fatal("insecure URL accepted")
	}
}
