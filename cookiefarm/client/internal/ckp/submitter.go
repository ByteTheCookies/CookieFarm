package ckp

import (
	"logger"
	"net"
	"server/database"

	"client/config"
)

func Start(flagsChan <-chan database.Flag) {
	logger.Log.Debug().Msg("Starting submission loop to the cookiefarm server with ckp ...")

	cm := config.GetInstance()
	addr := net.JoinHostPort(cm.GetHost(), "7777")

	conn, err := NewClient(addr)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Error connecting to ckp")
	}

	defer conn.Close()
	go conn.ReadPump(addr)

	for flag := range flagsChan {
		flagBytes := buildPayload(flag)
		err := conn.SendWithRetry(addr, flagBytes, 3)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Error sending flag to CKP server")
		}
	}

	logger.Log.Info().Msg("Submission loop finished")
}
