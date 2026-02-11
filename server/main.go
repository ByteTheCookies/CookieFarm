// Package main is the entry point for the API server.
package main

import (
	_ "embed"
	"logger"
	"server/cmd"

	_ "go.uber.org/automaxprocs"
)

func main() {
	if !logger.IsCompletionCommand() {
		logger.PrintBanner(true, "server")
	}
	cmd.Execute()
}
