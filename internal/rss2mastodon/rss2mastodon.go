// Package rss2mastodon provides the main logic for monitoring RSS feeds and posting updates to Mastodon.
// It handles configuration, feed checking, post processing, and integration with other components.
package rss2mastodon

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/toozej/rss2mastodon/internal/db"
	"github.com/toozej/rss2mastodon/internal/mastodon"
	"github.com/toozej/rss2mastodon/internal/rss"
	"github.com/toozej/rss2mastodon/pkg/config"
)

func Run(conf config.Config) {
	// Run starts the RSS to Mastodon monitoring loop. It initializes the database,
	// fetches RSS posts at regular intervals, filters them if a category is specified,
	// and handles posting new or updated posts to Mastodon.

	if conf.FeedURL == "" {
		log.Fatal("RSS feed URL is required")
	}

	if conf.Interval <= 0 {
		log.Error("Interval must be a positive integer")
		conf.Interval = 60 // Use default to prevent infinite loop
	}

	db.InitDB() // Initialize SQLite database
	defer db.CloseDB()

	for {
		posts, err := rss.CheckRSSFeed(conf.FeedURL)
		if err != nil {
			log.Printf("Error fetching RSS feed: %v", err)
			continue
		}

		for _, post := range posts {
			if conf.Category != "" {
				// Extract last segment of URL
				lastSegment := path.Base(post.Link)
				if !strings.Contains(lastSegment, conf.Category) {
					log.Debugf("Skipping post %s: category filter '%s' not in URL segment '%s'", post.Title, conf.Category, lastSegment)
					continue
				}
			}
			handlePost(post, &conf)
		}

		// Sleep for the configured interval before checking again
		time.Sleep(time.Duration(conf.Interval) * time.Minute)
	}
}

func handlePost(post rss.RSSItem, conf *config.Config) {
	// handlePost processes an RSS item, checks if it needs to be posted or updated on Mastodon,
	// sends the toot if necessary, and stores the post in the database.
	exists, updated, err := db.HasPostChanged(post.Link, post.Content)
	if err != nil {
		log.Error("Database error: ", err)
		return
	}

	var tootContent string
	var isUpdate bool

	switch {
	case exists && updated:
		// Post exists but is updated
		log.Printf("Post has been updated: %s", post.Title)
		tootContent = fmt.Sprintf("Blog post has been updated: %s", post.Link)
		isUpdate = true
	case !exists:
		// New post
		tootContent = mastodon.GetTootContent(post)
		isUpdate = false
	default:
		// Post exists but unchanged
		return
	}

	err = mastodon.TootPost(conf.MastodonURL, conf.MastodonAccessToken, tootContent)
	if err != nil {
		if isUpdate {
			log.Error("Failed to toot updated post: ", err)
		} else {
			log.Printf("Failed to toot new post: %v", err)
		}
		return
	}

	// Store the current content after successful toot
	err = db.StoreTootedPost(post.Link, post.Content)
	if err != nil {
		log.Error("Storing post toot in database failed: ", err)
	}
}
