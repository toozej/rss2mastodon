// Package main contains the entry point for the rss2mastodon application.
// It imports and executes the command-line interface from the cmd package.
package main

import cmd "github.com/toozej/rss2mastodon/cmd/rss2mastodon"

func main() {
	cmd.Execute()
}
