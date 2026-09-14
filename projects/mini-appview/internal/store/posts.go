package store

import (
	"database/sql"
	"time"
)

type Post struct {
	URI       string
	CID       string
	AuthorDID string
	Text      string
	CreatedAt time.Time
}

func SavePost(db *sql.DB, post Post) error {
	_, err := db.Exec(
		`INSERT INTO posts(
		uri,
		cid,
		author_did,
		text,
		created_at)
		VALUES (?, ?, ?, ?, ?)

		ON DUPLICATE KEY UPDATE
			cid = VALUES(cid),
			text = VALUES(text),
			created_at = VALUES(created_at)
			`,
		post.URI,
		post.CID,
		post.AuthorDID,
		post.Text,
		post.CreatedAt,
	)
	return err
}
