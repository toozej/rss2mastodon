package mastodon

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/toozej/rss2mastodon/internal/rss"

	"github.com/spf13/viper"
)

// GetTootContent constructs the toot message depending on the post title
func GetTootContent(post rss.RSSItem) string {
	if strings.HasPrefix(post.Title, "Thoughts") {
		return fmt.Sprintf("%s - %s", post.Content, post.Link)
	}
	return fmt.Sprintf("New blog post: %s", post.Link)
}

// TootPost sends a post to Mastodon
func TootPost(content string) error {
	mastodonURL := viper.GetString("mastodon_url")
	mastodonToken := viper.GetString("mastodon_token")

	if mastodonURL == "" || mastodonToken == "" {
		return fmt.Errorf("mastodon URL and token must be set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	formData := fmt.Sprintf("status=%s", content)
	req, err := http.NewRequest("POST", mastodonURL+"/api/v1/statuses", strings.NewReader(formData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mastodonToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	return nil
}
