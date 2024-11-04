package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"

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
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return
	}
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	_, err := maxprocs.Set()
	if err != nil {
		log.Error("Error setting maxprocs: ", err)
	}

	// create rootCmd-level flags
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug-level logging")
	rootCmd.Flags().StringP("feed-url", "f", "", "RSS feed URL to watch")
	rootCmd.Flags().IntP("interval", "i", 60, "Interval in minutes to check the RSS feed")

	// add sub-commands
	rootCmd.AddCommand(
		man.NewManCmd(),
		version.Command(),
	)
}
