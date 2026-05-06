package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"logger"
	"os"
	"strings"

	"client/api"
	"client/config"

	"github.com/spf13/cobra"
)

func configCheck(cm *config.ConfigManager) error {
	remoteCfg, err := api.GetConfig()
	if err != nil {
		return fmt.Errorf("error receiving shared configs: %w", err)
	}

	remoteConfigStr, err := json.Marshal(remoteCfg)
	if err != nil {
		return err
	}

	currentShc, err := json.Marshal(cm.Get().Shared)
	if err != nil {
		return err
	}

	if strings.TrimSpace(string(currentShc)) != strings.TrimSpace(string(remoteConfigStr)) {
		return errors.New("shared configuration has changed, update the config doing `ckc login`")
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "ckc",
	Short: "CLI client for interacting with the CookieFarm exploit framework",
	Long:  `CookieFarm is an automated exploitation framework developed by the ByteTheCookies team for the CyberChallenge competition. This CLI client connects to the CookieFarm server to deploy and manage exploits against target teams. To launch the terminal-based user interface (TUI), simply run the command "ckc" without any arguments.`, //nolint:revive
}

func syncConfig() {
	cm := config.GetInstance()
	cm.Read()
	checkErr := configCheck(cm)
	if checkErr != nil {
		if strings.Contains(checkErr.Error(), "Shared configuration has changed") {
			fmt.Fprintf(os.Stderr, "\n\033[1;33m[!] %v\033[0m\n", checkErr)
		} else {
			if strings.Contains(checkErr.Error(), "connect: connection refused") {
				fmt.Fprintf(os.Stderr, "\n\033[1;31m[+] Error connecting to server: %v\033[0m\n", checkErr)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "\n\033[1;31m[+] Error checking configuration: %v\033[0m\n", checkErr)
		}
	}
}

func buildCmd(useTUI *bool) *cobra.Command {
	var debug bool
	var useBanner bool

	if len(os.Args) != 1 {
		*useTUI = false
	}

	rootCmd.AddCommand(buildConfigCmd())
	rootCmd.AddCommand(buildExploitCmd())

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&useBanner, "banner", "B", false, "Show banner on startup")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if !logger.IsCompletionCommand() {
			logger.PrintBanner(!useBanner, "client")
		}

		if debug {
			logger.Setup("debug", true)
		} else {
			logger.Setup("info", true)
		}

		logger.Log.Debug().Msgf("Running command: %s", cmd.CalledAs())
		if cmd.CalledAs() == "completion" || cmd.CalledAs() == "help" || cmd.CalledAs() == "edit" {
			return
		}

		syncConfig()
	}

	return rootCmd
}
