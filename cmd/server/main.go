// Package main is the entry point for the API server.
package main

import (
	_ "embed"

	"github.com/ByteTheCookies/CookieFarm/cmd/server/cmd"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	_ "go.uber.org/automaxprocs"
)

func main() {
	if !logger.IsCompletionCommand() {
		logger.PrintBanner(true, "server")
	}
	cmd.Execute()
}
