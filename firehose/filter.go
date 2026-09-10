package main

import "strings"

type RecordType string

const (
	RecordPost    RecordType = "post"
	RecordLike    RecordType = "like"
	RecordFollow  RecordType = "follow"
	RecordProfile RecordType = "profile"
	RecordUnknown RecordType = "unknown"
)

func RecordTypeFromPath(path string) RecordType {
	collection := path

	if i := strings.Index(path, "/"); i != -1 {
		collection = path[:i]
	}

	switch collection {
	case "app.bsky.feed.post":
		return RecordPost

	case "app.bsky.feed.like":
		return RecordLike

	case "app.bsky.graph.follow":
		return RecordFollow

	case "app.bsky.actor.profile":
		return RecordProfile

	default:
		return RecordUnknown
	}
}
