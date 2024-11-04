package rss2mastodon

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Get environment variables
func getEnvVars() error {
	if _, err := os.Stat(".env"); err == nil {
		// Initialize Viper from .env file
		viper.SetConfigFile(".env") // Specify the name of your .env file

		// Read the .env file
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	// Enable reading environment variables
	viper.AutomaticEnv()

	// get mastodon_url from Viper
	mastodon_url := viper.GetString("MASTODON_URL")
	if mastodon_url == "" {
		return fmt.Errorf("mastodon_url must be provided")
	}

	mastodon_token := viper.GetString("MASTODON_TOKEN")
	if mastodon_token == "" {
		return fmt.Errorf("mastodon_token must be provided")
	}

	return nil
}
