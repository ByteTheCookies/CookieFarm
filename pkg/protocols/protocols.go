package protocols

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"plugin"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
)

type SubmitFunc = func(string, string, []string) ([]ResponseProtocol, error)

func LoadProtocol(protocolName string) (SubmitFunc, error) {
	exePath, err := os.Executable()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to get executable path")
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	pluginPath := path.Join(filepath.Dir(exePath), "..", "protocols", protocolName+".so")

	logger.Log.Debug().Str("plugin", pluginPath).Msg("Loading protocol plugin")

	plug, err := plugin.Open(pluginPath)
	if err != nil {
		logger.Log.Error().Err(err).Str("plugin", pluginPath).Msg("Failed to open plugin")
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	submitSymbol, err := plug.Lookup("Submit")
	if err != nil {
		logger.Log.Error().Err(err).Str("plugin", pluginPath).Msg("Submit symbol not found")
		return nil, fmt.Errorf("failed to lookup 'Submit': %w", err)
	}

	submitFunc, ok := submitSymbol.(SubmitFunc)
	if !ok {
		logger.Log.Error().Str("plugin", pluginPath).Msg("Invalid Submit function signature")
		return nil, errors.New("plugin 'Submit' has invalid signature")
	}

	logger.Log.Info().Str("protocol", protocolName).Msg("Protocol loaded successfully")

	return submitFunc, nil
}
