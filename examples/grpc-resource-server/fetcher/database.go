package fetcher

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"log/slog"
)

//go:embed books.json
var seedData []byte

type rawBook struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func EnsureDBSchema(db *sql.DB) error {
	slog.Info("ensuring database schema")
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS books (
  internal_id INTEGER PRIMARY KEY,
  id VARCHAR(14) NOT NULL UNIQUE,
  title TEXT NOT NULL
);
	`)

	return err
}

func SeedDB(db *sql.DB) error {
	slog.Info("seeding database")
	var books []rawBook

	err := json.Unmarshal(seedData, &books)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, book := range books {
		slog.Debug("adding book record", "id", book.ID, "title", book.Title)
		_, err = tx.Exec(`
INSERT INTO books (id, title) VALUES (?, ?)
		`, book.ID, book.Title)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
