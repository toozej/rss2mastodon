// Package config provides secure configuration management for the rss2mastodon application.
//
// This package handles loading configuration from environment variables and .env files
// with built-in security measures to prevent path traversal attacks. It uses the
// github.com/caarlos0/env library for environment variable parsing and
// github.com/joho/godotenv for .env file loading.
//
// The configuration loading follows a priority order:
//  1. Environment variables (highest priority)
//  2. .env file in current working directory
//  3. Default values (if any)
//
// Security features:
//   - Path traversal protection for .env file loading
//   - Secure file path resolution using filepath.Abs and filepath.Rel
//   - Validation against directory traversal attempts
//
// Example usage:
//
//	import "github.com/toozej/rss2mastodon/pkg/config"
//
//	func main() {
//		conf := config.GetEnvVars()
//		fmt.Printf("Mastodon URL: %s\n", conf.MastodonURL)
//	}
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config represents the application configuration structure.
//
// This struct defines all configurable parameters for the rss2mastodon
// application. Fields are tagged with struct tags that correspond to
// environment variable names for automatic parsing.
//
// Currently supported configuration:
//   - MastodonURL: Mastodon instance URL
//   - MastodonAccessToken: Access token for Mastodon API
//   - GotifyURL: Gotify instance URL
//   - GotifyToken: Token for Gotify notifications
//   - Debug: Enable debug logging
//   - FeedURL: RSS feed URL to watch
//   - Interval: Check interval in minutes (default 60)
type Config struct {
	// MastodonURL is the URL of the Mastodon instance.
	MastodonURL string `env:"MASTODON_URL"`

	// MastodonAccessToken is the access token for Mastodon API.
	MastodonAccessToken string `env:"MASTODON_ACCESS_TOKEN"`

	// GotifyURL is the URL of the Gotify instance.
	GotifyURL string `env:"GOTIFY_URL"`

	// GotifyToken is the token for Gotify notifications.
	GotifyToken string `env:"GOTIFY_TOKEN"`

	// Debug enables debug-level logging.
	Debug bool `env:"DEBUG"`

	// FeedURL is the RSS feed URL to watch.
	FeedURL string `env:"FEED_URL"`

	// Interval is the check interval in minutes.
	Interval int `env:"INTERVAL" envDefault:"60"`

	// Category is the URL category filter (optional).
	Category string `env:"CATEGORY"`
}

// GetEnvVars loads and returns the application configuration from environment
// variables and .env files with comprehensive security validation.
//
// This function performs the following operations:
//  1. Securely determines the current working directory
//  2. Constructs and validates the .env file path to prevent traversal attacks
//  3. Loads .env file if it exists in the current directory
//  4. Parses environment variables into the Config struct
//  5. Validates required fields
//  6. Returns the populated configuration or error
//
// Security measures implemented:
//   - Path traversal detection and prevention using filepath.Rel
//   - Absolute path resolution for secure path operations
//   - Validation against ".." sequences in relative paths
//   - Safe file existence checking before loading
//
// Errors returned for:
//   - Current directory access failures
//   - Path traversal attempts detected
//   - .env file parsing errors
//   - Environment variable parsing failures
//   - Missing required configuration
//
// Returns:
//   - Config: A populated configuration struct with values from environment
//     variables and/or .env file
//   - error: Any error encountered during loading or validation
//
// Example:
//
//	conf, err := config.GetEnvVars()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Mastodon URL: %s\n", conf.MastodonURL)
func GetEnvVars() (Config, error) {
	// Get current working directory for secure file operations
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("error getting current working directory: %w", err)
	}

	// Construct secure path for .env file within current directory
	envPath := filepath.Join(cwd, ".env")

	// Ensure the path is within our expected directory (prevent traversal)
	cleanEnvPath, err := filepath.Abs(envPath)
	if err != nil {
		return Config{}, fmt.Errorf("error resolving .env file path: %w", err)
	}
	cleanCwd, err := filepath.Abs(cwd)
	if err != nil {
		return Config{}, fmt.Errorf("error resolving current directory: %w", err)
	}
	relPath, err := filepath.Rel(cleanCwd, cleanEnvPath)
	if err != nil || strings.Contains(relPath, "..") {
		return Config{}, fmt.Errorf("error: .env file path traversal detected")
	}

	// Load .env file if it exists
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return Config{}, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Parse environment variables into config struct
	var conf Config
	if err := env.Parse(&conf); err != nil {
		return Config{}, fmt.Errorf("error parsing environment variables: %w", err)
	}

	// Validate required configuration
	if conf.MastodonURL == "" {
		return Config{}, fmt.Errorf("MASTODON_URL must be provided in .env file or environment")
	}
	if conf.MastodonAccessToken == "" {
		return Config{}, fmt.Errorf("MASTODON_ACCESS_TOKEN must be provided in .env file or environment")
	}
	if conf.GotifyURL == "" {
		return Config{}, fmt.Errorf("GOTIFY_URL must be provided in .env file or environment")
	}
	if conf.GotifyToken == "" {
		return Config{}, fmt.Errorf("GOTIFY_TOKEN must be provided in .env file or environment")
	}

	return conf, nil
}
