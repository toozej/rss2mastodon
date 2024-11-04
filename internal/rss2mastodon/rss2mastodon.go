package rss2mastodon

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/toozej/rss2mastodon/internal/db"
	"github.com/toozej/rss2mastodon/internal/mastodon"
	"github.com/toozej/rss2mastodon/internal/rss"
)

func Run(cmd *cobra.Command, args []string) {
	err := getEnvVars()
	if err != nil {
		log.Fatal("Error gathering required environment variables: ", err)
	}

	feedURL := viper.GetString("feed_url")
	if feedURL == "" {
		log.Fatal("RSS feed URL is required")
	}

	db.InitDB() // Initialize SQLite database
	defer db.CloseDB()

	// Get interval from environment variable or flag (default to 10 minutes)
	interval := viper.GetInt("interval")
	if interval <= 0 {
		log.Error("Interval must be a positive integer")
	}

	for {
		posts, err := rss.CheckRSSFeed(feedURL)
		if err != nil {
			log.Printf("Error fetching RSS feed: %v", err)
			continue
		}

		for _, post := range posts {
			handlePost(post)
		}

		// Sleep for the configured interval before checking again
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func handlePost(post rss.RSSItem) {
	exists, updated, err := db.HasPostChanged(post.Link, post.Content)
	if err != nil {
		log.Error("Database error: ", err)
		return
	}

	if exists && updated {
		// Post exists but is updated
		log.Printf("Post has been updated: %s", post.Title)
		tootContent := fmt.Sprintf("Blog post has been updated: %s", post.Link)
		err := mastodon.TootPost(tootContent)
		if err != nil {
			log.Error("Failed to toot updated post: ", err)
		} else {
			err = db.StoreTootedPost(post.Link, post.Content)
			if err != nil {
				log.Error("Storing updated post toot in database failed: ", err)
			}
		}
	} else if !exists {
		// New post
		tootContent := mastodon.GetTootContent(post)
		err := mastodon.TootPost(tootContent)
		if err != nil {
			log.Printf("Failed to toot new post: %v", err)
		} else {
			err = db.StoreTootedPost(post.Link, post.Content)
			if err != nil {
				log.Error("Storing new post toot in database failed: ", err)
			}
		}
	}
}
