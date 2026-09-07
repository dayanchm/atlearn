package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// Session holds credentials returned by the PDS. Do not log this value.
type Session struct {
	AccessJWT  string `json:"accessJwt"`
	RefreshJWT string `json:"refreshJwt"`
	DID        string `json:"did"`
	Handle     string `json:"handle"`
	service    string
}

type PostResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

var postHTTPClient = &http.Client{
	Timeout: 20 * time.Second,

	// Keep credentials and writes on the explicitly configured server.
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// CreateSession logs in with an app password.
// BSKY_PDS_URL defaults to bsky.social.
func CreateSession(handle, password string) (*Session, error) {
	handle = strings.TrimSpace(handle)

	if handle == "" || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("handle and app password are required")
	}

	service := strings.TrimRight(
		strings.TrimSpace(os.Getenv("BSKY_PDS_URL")),
		"/",
	)

	if service == "" {
		service = "https://bsky.social"
	}

	u, err := url.Parse(service)
	if err != nil ||
		u.Scheme != "https" ||
		u.Host == "" ||
		u.User != nil ||
		u.RawQuery != "" ||
		u.Fragment != "" {
		return nil, fmt.Errorf("BSKY_PDS_URL must be an HTTPS server URL")
	}

	var session Session

	if err := postXRPC(
		service,
		"com.atproto.server.createSession",
		"",
		SessionRequest{
			Identifier: handle,
			Password:   password,
		},
		&session,
	); err != nil {
		return nil, err
	}

	if session.AccessJWT == "" || session.DID == "" {
		return nil, fmt.Errorf(
			"session response is missing accessJwt or did",
		)
	}

	session.service = service

	return &session, nil
}

// CreatePost publishes one text post.
func CreatePost(session *Session, text string) (*PostResponse, error) {
	return createPostWithKey(session, text, "")
}

func createPostWithKey(
	session *Session,
	text string,
	rkey string,
) (*PostResponse, error) {

	if session == nil ||
		session.AccessJWT == "" ||
		session.DID == "" ||
		session.service == "" {

		return nil, fmt.Errorf(
			"a session from CreateSession is required",
		)
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("post text must not be empty")
	}

	if !utf8.ValidString(text) {
		return nil, fmt.Errorf("post text must be valid UTF-8")
	}

	// Bluesky post limit.
	if utf8.RuneCountInString(text) > 300 {
		return nil, fmt.Errorf(
			"post text exceeds Bluesky's 300-character limit",
		)
	}

	request := struct {
		RKey       string `json:"rkey,omitempty"`
		Repo       string `json:"repo"`
		Collection string `json:"collection"`

		Record struct {
			Type      string `json:"$type"`
			Text      string `json:"text"`
			CreatedAt string `json:"createdAt"`
		} `json:"record"`
	}{
		RKey:       rkey,
		Repo:       session.DID,
		Collection: "app.bsky.feed.post",
	}

	request.Record.Type = "app.bsky.feed.post"
	request.Record.Text = text

	// Use standard RFC3339 timestamp.
	request.Record.CreatedAt = time.Now().
		UTC().
		Format(time.RFC3339)

	var result PostResponse

	if err := postXRPC(
		session.service,
		"com.atproto.repo.createRecord",
		session.AccessJWT,
		request,
		&result,
	); err != nil {
		return nil, err
	}

	if result.URI == "" || result.CID == "" {
		return nil, fmt.Errorf(
			"post response is missing uri or cid; verify the account before retrying",
		)
	}

	return &result, nil
}

// PostPoetry publishes the title, first four nonempty lines and author.
// Poems shorter than four lines use all available lines.
func PostPoetry(
	session *Session,
	poem Poetry,
) (*PostResponse, error) {

	if strings.TrimSpace(poem.Title) == "" ||
		strings.TrimSpace(poem.Text) == "" ||
		strings.TrimSpace(poem.Author) == "" {

		return nil, fmt.Errorf(
			"poem title, text and author are required",
		)
	}

	return CreatePost(
		session,
		poemPostText(poem),
	)
}

func postXRPC(
	service string,
	method string,
	token string,
	input any,
	output any,
) error {

	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf(
			"encode %s request: %w",
			method,
			err,
		)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		service+"/xrpc/"+method,
		bytes.NewReader(body),
	)

	if err != nil {
		return fmt.Errorf(
			"create %s request: %w",
			method,
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Accept",
		"application/json",
	)

	if token != "" {
		req.Header.Set(
			"Authorization",
			"Bearer "+token,
		)
	}

	resp, err := postHTTPClient.Do(req)

	if err != nil {
		return fmt.Errorf(
			"%s request failed (verify account before retrying a post): %w",
			method,
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {

		var apiError struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}

		_ = json.NewDecoder(
			io.LimitReader(resp.Body, 1<<20),
		).Decode(&apiError)

		return fmt.Errorf(
			"%s: HTTP %d (%s): %s",
			method,
			resp.StatusCode,
			apiError.Error,
			apiError.Message,
		)
	}

	if err := json.NewDecoder(
		io.LimitReader(resp.Body, 1<<20),
	).Decode(output); err != nil {

		return fmt.Errorf(
			"decode %s response (verify account before retrying a post): %w",
			method,
			err,
		)
	}

	return nil
}
