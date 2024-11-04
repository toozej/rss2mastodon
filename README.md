# rss2mastodon

rss2mastodon is a CLI tool that monitors an RSS feed for new posts and automatically posts updates to a specified Mastodon instance. This application is designed for easy configuration and seamless integration, making it simple to announce new blog posts or content updates on your Mastodon account.

## Features
- Periodically checks an RSS feed for new or updated posts.
- Posts updates to a configured Mastodon server.
- Stores previously tooted posts in an SQLite database to avoid reposting.
- Configurable check interval and customizable toot content.
- Debug mode for more detailed logging.

## Installation
### Prerequisites
- Go (version 1.17 or later)
- SQLite for database management
- Make

### Steps
1.	Clone the repository:
```bash
git clone https://github.com/toozej/rss2mastodon.git
cd rss2mastodon
```

2.	Build the executable:
`make build`

## Usage
1.	Set Environment Variables:
    Create a .env file in the root of your project or set the required environment variables directly:

    ```
    MASTODON_URL=https://your-mastodon-instance
    MASTODON_TOKEN=your-access-token
    FEED_URL=https://example.com/rss
    ```

    Alternatively, you can provide the feed-url and interval as command-line flags or environment variables.
2.	Run the application:
    ```bash
    ./rss2mastodon --feed-url "https://example.com/rss" --interval 60
    ```

    `--feed-url`: The URL of the RSS feed to monitor.
    `--interval`: The interval in minutes for checking the RSS feed (default is 60 minutes).

3. Enable Debug Mode:
Use the --debug flag to enable debug-level logging for troubleshooting.
```bash
./rss2mastodon --debug
```


## Major Components
### Command Structure (cmd/rss2mastodon/root.go)
- Defines the main rss2mastodon command and its subcommands (man and version).
- Sets up CLI flags and binds them to configuration via Viper.

### Configuration (internal/rss2mastodon/config.go)
- Loads configuration from environment variables and the .env file if present.
- Ensures required variables (MASTODON_URL, MASTODON_TOKEN) are set.

### RSS Handling (internal/rss/rss.go)
- Fetches and parses the RSS feed.
- Provides hashing functionality to detect changes in post content.

### Mastodon Integration (internal/mastodon/mastodon.go)
- Constructs toot content based on the post title and content.
- Sends HTTP requests to post updates on the Mastodon instance.

### Database Management (internal/db/db.go)
- Manages an SQLite database to store and check previously tooted posts.
- Functions for initializing the database, storing, and verifying post changes.

## update golang version
- `make update-golang-version`
