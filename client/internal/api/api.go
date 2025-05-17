// Package api provides functions to interact with the CookieFarm server API.
package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ByteTheCookies/cookieclient/internal/config"
	"github.com/ByteTheCookies/cookieclient/internal/logger"
	"github.com/ByteTheCookies/cookieclient/internal/models"
	json "github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

var client = &http.Client{}

// SendFlag sends flags to the CookieFarm server API.
func SendFlag(flags ...models.Flag) error {
	body, err := json.Marshal(map[string][]models.Flag{"flags": flags})
	if err != nil {
		logger.Log.Error().Err(err).Msg("error during flags marshalling")
		return err
	}

	ServerURL := "http://" + *config.HostServer + "/api/v1/submit-flags"
	req, err := http.NewRequest(http.MethodPost, ServerURL, bytes.NewReader(body))
	if err != nil {
		log.Error().Err(err).Str("url", ServerURL).Msg("error creating request")
		return err
	}

	req.Header.Set("Cookie", "token="+config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Str("url", ServerURL).Msg("error during flags submission request")
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error().Err(err).Msg("error reading response")
		return err
	}

	logger.Log.Debug().
		Int("status", resp.StatusCode).
		Msgf("Server response from submit-flags: %s", string(respBody))

	return nil
}

// GetConfig retrieves the configuration from the CookieFarm server API.
func GetConfig() (models.Config, error) {
	ServerURL := "http://" + *config.HostServer + "/api/v1/config"
	req, err := http.NewRequest(http.MethodGet, ServerURL, nil)
	if err != nil {
		return models.Config{}, fmt.Errorf("error creating config request: %w", err)
	}
	req.Header.Set("Cookie", "token="+config.Token)

	resp, err := client.Do(req)
	if err != nil {
		return models.Config{}, fmt.Errorf("error sending config request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Config{}, fmt.Errorf("error reading config response: %w", err)
	}

	var parsedConfig models.Config
	if err := json.Unmarshal(respBody, &parsedConfig); err != nil {
		return models.Config{}, fmt.Errorf("error parsing config: %w", err)
	}

	logger.Log.Debug().Msgf("Configuration received correctly")

	return parsedConfig, nil
}

// Login sends a login request to the CookieFarm server API.
func Login(password string) (string, error) {
	ServerURL := "http://" + *config.HostServer + "/api/v1/auth/login"

	_, err := url.Parse(ServerURL)
	if err != nil {
		log.Fatal().Msg("Invalid base URL in config")
	}

	logger.Log.Debug().Str("url", ServerURL).Msg("Login attempt")

	resp, err := http.Post(
		ServerURL,
		"application/x-www-form-urlencoded",
		bytes.NewBufferString("password="+password),
	)
	if err != nil {
		logger.Log.Error().Err(err).Msg("error sending login request")
		return "", err
	}
	defer resp.Body.Close()

	cookies := resp.Cookies()
	for _, c := range cookies {
		if c.Name == "token" {
			logger.Log.Debug().Str("token", c.Value).Msg("Token found")
			logger.Log.Info().Msg("Login successfully")
			return c.Value, nil
		}
	}

	logger.Log.Warn().Msg("Token not found in Set-Cookie")
	return "", errors.New("token not found in Set-Cookie")
}
