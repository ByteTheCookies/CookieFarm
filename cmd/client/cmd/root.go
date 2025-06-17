package cmd

import (
	"os"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	debug     bool
	useTUI    bool
	useBanner bool
)

// RootCmd represents the base command when called without any subcommands
// Exported for TUI usage
var RootCmd = &cobra.Command{
	Use:   "ckc",
	Short: "The client cli for CookieFarm",
	Long: `CookieFarm is a exploiter writed by the team ByteTheCookies for CyberChallenge
	competition. This is the client cli for the CookieFarm server for attack the teams with exploits.`, // Da migliorare
	Version: "v1.1.0",
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(ConfigCmd)
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")
	RootCmd.PersistentFlags().BoolVarP(&useTUI, "no-tui", "N", false, "Disable TUI mode")
	RootCmd.PersistentFlags().BoolVarP(&useBanner, "no-banner", "B", false, "Remove banner on startup")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cm := config.GetConfigManager()
		cm.SetUseBanner(useBanner)
		if debug {
			logger.Setup("debug", true)
		} else {
			logger.Setup("info", true)
		}
	}
}
