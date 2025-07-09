package cmd

import (
	"context"
	"os"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
// Exported for TUI usage
var RootCmd = &cobra.Command{
	Use:     "ckc",
	Short:   "CLI client for interacting with the CookieFarm exploit framework",
	Long:    `CookieFarm is an automated exploitation framework developed by the ByteTheCookies team for the CyberChallenge competition. This CLI client connects to the CookieFarm server to deploy and manage exploits against target teams. To launch the terminal-based user interface (TUI), simply run the command "ckc" without any arguments.`, //nolint:revive
	Version: models.VERSION,
}

func Execute() {
	theme := logger.CookieCLIColorSchema
	if err := fang.Execute(context.TODO(), RootCmd, fang.WithVersion(models.VERSION), fang.WithTheme(theme)); err != nil {
		os.Exit(1)
	}
}

func init() {
	var debug, useBanner bool

	RootCmd.AddCommand(ConfigCmd)
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")
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
