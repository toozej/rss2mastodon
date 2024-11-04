package rss

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	Content string `xml:"description"`
}

// CheckRSSFeed fetches and parses the RSS feed from the provided URL
func CheckRSSFeed(feedURL string) ([]RSSItem, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	var feed RSSFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	return feed.Channel.Items, nil
}

// HashContent creates a SHA-256 hash of the post content
func HashContent(content string) [32]byte {
	return sha256.Sum256([]byte(content))
}
