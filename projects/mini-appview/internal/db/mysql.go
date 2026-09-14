package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func Open() (*sql.DB, error) {

	dsn := "root:@tcp(127.0.0.1:3306)/"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE DATABASE IF NOT EXISTS appview
		CHARACTER SET utf8mb4
		COLLATE utf8mb4_unicode_ci
	`); err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}
	db.Close()

	dsn = "root:@tcp(127.0.0.1:3306)/appview?parseTime=true"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			uri VARCHAR(512) PRIMARY KEY,
			cid VARCHAR(255) NOT NULL,
			author_did VARCHAR(255) NOT NULL,
			text TEXT,
			created_at DATETIME(6),
			indexed_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
		)
	`)
	if err != nil {
		return fmt.Errorf("create posts table: %w", err)
	}
	return nil

}
