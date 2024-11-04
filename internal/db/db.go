package db

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
	"github.com/toozej/rss2mastodon/internal/rss"
)

var db *sql.DB

// InitDB initializes the SQLite database
func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "./tooted_posts.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Create table if not exists
	query := `CREATE TABLE IF NOT EXISTS tooted_posts (
		link TEXT PRIMARY KEY,
		content_hash TEXT,
		timestamp TEXT
	)`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

// CloseDB closes the SQLite database connection
func CloseDB() {
	err := db.Close()
	if err != nil {
		log.Error("Error closing SQLite database connection: ", err)
	}
}

// StoreTootedPost stores the link, content hash, and timestamp in the database
func StoreTootedPost(link string, content string) error {
	query := `INSERT OR REPLACE INTO tooted_posts(link, content_hash, timestamp) VALUES (?, ?, ?)`
	contentHash := rss.HashContent(content)
	_, err := db.Exec(query, link, fmt.Sprintf("%x", contentHash), time.Now().Format(time.RFC3339))
	return err
}

// HasPostChanged checks if the post content has changed or if it is new
func HasPostChanged(link string, content string) (exists bool, updated bool, err error) {
	query := `SELECT content_hash FROM tooted_posts WHERE link = ?`
	row := db.QueryRow(query, link)

	var storedHash string
	err = row.Scan(&storedHash)
	if err == sql.ErrNoRows {
		// Post is new
		return false, false, nil
	} else if err != nil {
		return false, false, err
	}

	// Check if the content hash has changed
	newHash := fmt.Sprintf("%x", rss.HashContent(content))
	if storedHash != newHash {
		// Post has been updated
		return true, true, nil
	}

	// Post already exists and is unchanged
	return true, false, nil
}
