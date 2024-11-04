#!/bin/sh
set -e
rm -rf manpages
mkdir manpages
go run ./cmd/rss2mastodon/ man | gzip -c -9 >manpages/rss2mastodon.1.gz
