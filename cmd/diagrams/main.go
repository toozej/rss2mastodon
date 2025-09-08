// Package main provides diagram generation utilities for the rss2mastodon project.
//
// This application generates architectural and component diagrams for the rss2mastodon
// application using the go-diagrams library. It creates visual representations of the
// project structure and component relationships to aid in documentation and understanding.
//
// The generated diagrams are saved as .dot files in the docs/diagrams/go-diagrams/
// directory and can be converted to various image formats using Graphviz.
//
// Usage:
//
//	go run cmd/diagrams/main.go
//
// This will generate:
//   - architecture.dot: High-level architecture showing RSS feed monitoring flow
//   - components.dot: Component relationships and dependencies
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/generic"
	"github.com/blushft/go-diagrams/nodes/programming"
)

// main is the entry point for the diagram generation utility.
//
// This function orchestrates the entire diagram generation process:
//  1. Creates the output directory structure
//  2. Changes to the appropriate working directory
//  3. Generates architecture and component diagrams
//  4. Reports successful completion
//
// The function will terminate with log.Fatal if any critical operation fails,
// such as directory creation, navigation, or diagram rendering.
func main() {
	// Ensure output directory exists
	if err := os.MkdirAll("docs/diagrams", 0750); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Change to docs/diagrams directory
	if err := os.Chdir("docs/diagrams"); err != nil {
		log.Fatal("Failed to change directory:", err)
	}

	// Generate architecture diagram
	generateArchitectureDiagram()

	// Generate component diagram
	generateComponentDiagram()

	fmt.Println("Diagram .dot files generated successfully in ./docs/diagrams/go-diagrams/")
}

// generateArchitectureDiagram creates a high-level architecture diagram showing
// the RSS feed monitoring and Mastodon posting flow for the rss2mastodon application.
//
// The diagram illustrates:
//   - RSS feed monitoring and parsing
//   - Database storage for tracking posted items
//   - Mastodon API integration for posting
//   - Gotify notifications for status updates
//   - Configuration management flow
//
// The diagram is rendered in top-to-bottom (TB) direction and saved as
// "architecture.dot" in the current working directory. The function will
// terminate the program with log.Fatal if diagram creation or rendering fails.
func generateArchitectureDiagram() {
	d, err := diagram.New(diagram.Filename("architecture"), diagram.Label("RSS2Mastodon Architecture"), diagram.Direction("TB"))
	if err != nil {
		log.Fatal(err)
	}

	// Define components
	rssFeed := generic.Blank.Blank(diagram.NodeLabel("RSS Feed"))
	rssParser := programming.Language.Go(diagram.NodeLabel("RSS Parser"))
	database := generic.Blank.Blank(diagram.NodeLabel("SQLite Database"))
	mastodonAPI := generic.Blank.Blank(diagram.NodeLabel("Mastodon API"))
	gotifyAPI := generic.Blank.Blank(diagram.NodeLabel("Gotify API"))
	config := generic.Blank.Blank(diagram.NodeLabel("Configuration\n(env/godotenv)"))
	logging := generic.Blank.Blank(diagram.NodeLabel("Logging\n(logrus)"))

	// Create connections showing the flow
	d.Connect(rssFeed, rssParser, diagram.Forward())
	d.Connect(rssParser, database, diagram.Forward())
	d.Connect(rssParser, mastodonAPI, diagram.Forward())
	d.Connect(rssParser, gotifyAPI, diagram.Forward())
	d.Connect(config, rssParser, diagram.Forward())
	d.Connect(logging, rssParser, diagram.Forward())

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}

// generateComponentDiagram creates a detailed component diagram showing the
// relationships and dependencies between different packages in the rss2mastodon project.
//
// The diagram illustrates:
//   - main.go as the entry point
//   - cmd/rss2mastodon package handling CLI operations
//   - Integration with internal packages (db, rss, mastodon, gotify, rss2mastodon)
//   - Integration with pkg packages (config, version, man)
//   - Data flow between components
//
// The diagram is rendered in left-to-right (LR) direction and saved as
// "components.dot" in the current working directory. The function will
// terminate the program with log.Fatal if diagram creation or rendering fails.
func generateComponentDiagram() {
	d, err := diagram.New(diagram.Filename("components"), diagram.Label("RSS2Mastodon Components"), diagram.Direction("LR"))
	if err != nil {
		log.Fatal(err)
	}

	// Main components
	main := programming.Language.Go(diagram.NodeLabel("main.go"))
	rootCmd := programming.Language.Go(diagram.NodeLabel("cmd/rss2mastodon\nroot.go"))
	config := programming.Language.Go(diagram.NodeLabel("pkg/config\nconfig.go"))
	rss2mastodon := programming.Language.Go(diagram.NodeLabel("internal/rss2mastodon\nrss2mastodon.go"))

	// Internal packages
	db := programming.Language.Go(diagram.NodeLabel("internal/db\ndb.go"))
	rss := programming.Language.Go(diagram.NodeLabel("internal/rss\nrss.go"))
	mastodon := programming.Language.Go(diagram.NodeLabel("internal/mastodon\nmastodon.go"))
	gotify := programming.Language.Go(diagram.NodeLabel("internal/gotify\ngotify.go"))

	// Pkg packages
	version := programming.Language.Go(diagram.NodeLabel("pkg/version\nversion.go"))
	man := programming.Language.Go(diagram.NodeLabel("pkg/man\nman.go"))

	// Create connections showing the flow
	d.Connect(main, rootCmd, diagram.Forward())
	d.Connect(rootCmd, config, diagram.Forward())
	d.Connect(rootCmd, rss2mastodon, diagram.Forward())
	d.Connect(rootCmd, version, diagram.Forward())
	d.Connect(rootCmd, man, diagram.Forward())

	// Internal package connections
	d.Connect(rss2mastodon, db, diagram.Forward())
	d.Connect(rss2mastodon, rss, diagram.Forward())
	d.Connect(rss2mastodon, mastodon, diagram.Forward())
	d.Connect(rss2mastodon, gotify, diagram.Forward())

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}
