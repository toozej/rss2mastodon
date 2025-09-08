// Package cmd contains the command-line interface implementation for the rss2mastodon application.
// It defines the root Cobra command and its execution logic.
package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/toozej/rss2mastodon/internal/rss2mastodon"
	"github.com/toozej/rss2mastodon/pkg/man"
	"github.com/toozej/rss2mastodon/pkg/version"
)

var rootCmd = &cobra.Command{
	Use:              "rss2mastodon",
	Short:            "Watches a RSS feed for new posts, then announces them on Mastodon",
	Long:             `Watches a RSS feed for new posts, then announces them on Mastodon`,
	Args:             cobra.ExactArgs(0),
	PersistentPreRun: rootCmdPreRun,
	Run:              rss2mastodon.Run,
}

func rootCmdPreRun(cmd *cobra.Command, args []string) {
	// rootCmdPreRun is the persistent pre-run function for the root command.
	// It configures logging level based on the debug flag.
	debug, err := cmd.PersistentFlags().GetBool("debug")
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

func Execute() {
	// Execute runs the root command and handles any execution errors by printing them and exiting with status 1.
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")
	rootCmd.Flags().StringP("feed-url", "f", "", "RSS feed URL to watch")
	rootCmd.Flags().IntP("interval", "i", 60, "Interval in minutes to check the RSS feed")
	rootCmd.Flags().StringP("category", "c", "", "Category to filter URL last segment")

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
