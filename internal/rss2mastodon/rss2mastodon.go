// Package rss2mastodon provides the main logic for monitoring RSS feeds and posting updates to Mastodon.
// It handles configuration, feed checking, post processing, and integration with other components.
package rss2mastodon

import (
	"fmt"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/toozej/rss2mastodon/internal/db"
	"github.com/toozej/rss2mastodon/internal/mastodon"
	"github.com/toozej/rss2mastodon/internal/rss"
	"github.com/toozej/rss2mastodon/pkg/config"
)

func Run(cmd *cobra.Command, args []string) {
	// Run starts the RSS to Mastodon monitoring loop. It initializes the database,
	// fetches RSS posts at regular intervals, filters them if a category is specified,
	// and handles posting new or updated posts to Mastodon.
	conf, err := config.GetEnvVars()
	if err != nil {
		log.Fatal("Error gathering required environment variables: ", err)
	}

	// Override debug from CLI flag if set
	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		conf.Debug = true
	}

	feedURL := conf.FeedURL
	feedURLFlag, err := cmd.Flags().GetString("feed-url")
	if err != nil {
		log.Fatal(err)
	}
	if feedURLFlag != "" {
		feedURL = feedURLFlag
	}
	if feedURL == "" {
		log.Fatal("RSS feed URL is required")
	}

	db.InitDB() // Initialize SQLite database
	defer db.CloseDB()

	// Get interval from environment variable or flag (default to 60 minutes)
	interval := conf.Interval
	intervalFlag, err := cmd.Flags().GetInt("interval")
	if err != nil {
		log.Error(err)
	}
	if intervalFlag > 0 {
		interval = intervalFlag
	}
	if interval <= 0 {
		log.Error("Interval must be a positive integer")
		interval = 60 // Use default to prevent infinite loop
	}

	// Get category from environment variable or flag (optional)
	category := conf.Category
	categoryFlag, err := cmd.Flags().GetString("category")
	if err != nil {
		log.Error(err)
	}
	if categoryFlag != "" {
		category = categoryFlag
	}

	for {
		posts, err := rss.CheckRSSFeed(feedURL)
		if err != nil {
			log.Printf("Error fetching RSS feed: %v", err)
			continue
		}

		for _, post := range posts {
			if category != "" {
				// Extract last segment of URL
				lastSegment := path.Base(post.Link)
				if !strings.Contains(lastSegment, category) {
					log.Debugf("Skipping post %s: category filter '%s' not in URL segment '%s'", post.Title, category, lastSegment)
					continue
				}
			}
			handlePost(post, &conf)
		}

		// Sleep for the configured interval before checking again
		time.Sleep(time.Duration(interval) * time.Minute)
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
