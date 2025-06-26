package cmd

import (
	"context"
	"image/color"
	"os"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var (
	debug     bool
	useBanner bool
)

const VERSION = "v1.2.0"

// RootCmd represents the base command when called without any subcommands
// Exported for TUI usage
var RootCmd = &cobra.Command{
	Use:   "ckc",
	Short: "The client cli for CookieFarm",
	Long: `CookieFarm is a exploiter writed by the team ByteTheCookies for CyberChallenge
competition. This is the client cli for the CookieFarm server for attack the teams with exploits.`, // Da migliorare
	Version: VERSION,
}

func Execute() {
	theme := fang.ColorScheme{
		Base:           color.RGBA{0xe9, 0xe9, 0xe9, 0xe9},
		Title:          color.RGBA{0xCD, 0xA1, 0x57, 0xff},
		Description:    color.RGBA{0xD9, 0xD9, 0xD9, 0xff},
		Codeblock:      color.RGBA{0x0a, 0x0c, 0x0d, 0xff},
		Program:        color.RGBA{0xCD, 0xA1, 0x57, 0xff},
		DimmedArgument: color.RGBA{0x88, 0x88, 0x88, 0xff},
		Comment:        color.RGBA{0x88, 0x88, 0x88, 0xff},
		Flag:           color.RGBA{0x21, 0x96, 0xF3, 0xff},
		FlagDefault:    color.RGBA{0xD9, 0xD9, 0xD9, 0xff},
		Command:        color.RGBA{0xCD, 0xA1, 0x57, 0xff},
		QuotedString:   color.RGBA{0x21, 0x9B, 0x54, 0xff},
		Argument:       color.RGBA{0xED, 0xED, 0xED, 0xff},
		Help:           color.RGBA{0x88, 0x88, 0x88, 0xff},
		Dash:           color.RGBA{0x88, 0x88, 0x88, 0xff},
		ErrorHeader:    [2]color.Color{color.RGBA{0xED, 0xED, 0xED, 0xff}, color.RGBA{0xE7, 0x4C, 0x3C, 0xff}},
		ErrorDetails:   color.RGBA{0xE7, 0x4C, 0x3C, 0xff},
	}
	if err := fang.Execute(context.TODO(), RootCmd, fang.WithVersion(VERSION), fang.WithTheme(theme)); err != nil {
		os.Exit(1)
	}
}

func init() {
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
