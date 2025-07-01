// Package api provides functions to interact with the CookieFarm server API.
package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	json "github.com/bytedance/sonic"
)

func parseURL(host, port, endpoint string) (string, error) {
	URL := "http://" + host + ":" + port + endpoint
	_, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	return URL, nil
}

// GetConfig retrieves the configuration from the CookieFarm server API.
func GetConfig() (models.ConfigShared, error) {
	cm := config.GetConfigManager()
	localConfig := cm.GetLocalConfig()
	client := &http.Client{}

	serverURL, err := parseURL(localConfig.Host, strconv.Itoa(int(localConfig.Port)), "/api/v1/config")
	if err != nil {
		logger.Log.Error().Err(err).Msg("Invalid base URL in config")
	}

	req, err := http.NewRequest(http.MethodGet, serverURL, nil)
	if err != nil {
		return models.ConfigShared{}, fmt.Errorf("error creating config request: %w", err)
	}
	req.Header.Set("Cookie", "token="+cm.GetToken())

	resp, err := client.Do(req)
	if err != nil {
		return models.ConfigShared{}, fmt.Errorf("error sending config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msgf("Error fetching config: %s", body)
		return models.ConfigShared{}, fmt.Errorf("error fetching config: %s", body)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.ConfigShared{}, fmt.Errorf("error reading config response: %w", err)
	}

	var parsedConfig models.ConfigShared
	if err := json.Unmarshal(respBody, &parsedConfig); err != nil {
		return models.ConfigShared{}, fmt.Errorf("error parsing config: %w", err)
	}

	logger.Log.Debug().Msgf("Configuration received correctly")

	return parsedConfig, nil
}

// Login sends a login request to the CookieFarm server API.
func Login(password string) (string, error) {
	cm := config.GetConfigManager()
	localConfig := cm.GetLocalConfig()

	serverURL, err := parseURL(localConfig.Host, strconv.Itoa(int(localConfig.Port)), "/api/v1/auth/login")
	if err != nil {
		logger.Log.Error().Err(err).Msg("Invalid base URL in config")
	}

	logger.Log.Debug().Str("url", serverURL).Msg("Login attempt")

	resp, err := http.Post(
		serverURL,
		"application/x-www-form-urlencoded",
		bytes.NewBufferString("username="+cm.GetLocalConfig().Username+"&password="+password),
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

// SubmitBatchDirect sends a batch of flags directly to the CookieFarm server API.
//
// @IMPORTANT: I do not raccomend using this function, use websockets instead.
func SubmitBatchDirect(flags []models.ClientData) (string, error) {
	cm := config.GetConfigManager()
	localConfig := cm.GetLocalConfig()
	client := &http.Client{}

	serverURL, err := parseURL(localConfig.Host, strconv.Itoa(int(localConfig.Port)), "/api/v1/submit-flags-standalone")
	if err != nil {
		logger.Log.Error().Err(err).Msg("Invalid base URL in config")
	}
	logger.Log.Debug().Str("url", serverURL).Msg("Login attempt")

	flagMarshalled, err := json.Marshal(models.SubmitFlagsRequest{Flags: flags})
	if err != nil {
		logger.Log.Error().Err(err).Msg("error marshalling flags")
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, serverURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating config request: %w", err)
	}
	req.Header.Set("Cookie", "token="+cm.GetToken())
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(flagMarshalled))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msgf("Error submitting flags: %s", body)
		return "", fmt.Errorf("error submitting flags: %s", body)
	}

	if resp.StatusCode == http.StatusOK {
		logger.Log.Info().Msg("Flags submitted successfully")
		return "Flags submitted successfully", nil
	} else {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msgf("Unexpected response status: %d, body: %s", resp.StatusCode, body)
		return "", fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, body)
	}
}

// SubmitDirect sends a single flag directly to the CookieFarm server API.
func SubmitDirect(flag models.ClientData) (string, error) {
	cm := config.GetConfigManager()
	localConfig := cm.GetLocalConfig()
	client := &http.Client{}

	serverURL, err := parseURL(localConfig.Host, strconv.Itoa(int(localConfig.Port)), "/api/v1/submit-flag")
	if err != nil {
		logger.Log.Error().Err(err).Msg("Invalid base URL in config")
	}
	logger.Log.Debug().Str("url", serverURL).Msg("Login attempt")

	flagMarshalled, err := json.Marshal(models.SubmitFlagRequest{Flag: flag})
	if err != nil {
		logger.Log.Error().Err(err).Msg("error marshalling flags")
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, serverURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating config request: %w", err)
	}
	req.Header.Set("Cookie", "token="+cm.GetToken())
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(flagMarshalled))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending config request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msgf("Error submitting flags: %s", body)
		return "", fmt.Errorf("error submitting flags: %s", body)
	}

	logger.Log.Debug().Msgf("Response status code: %d", resp.StatusCode)
	logger.Log.Debug().Msgf("Response headers: %v", resp.Header)
	logger.Log.Debug().Msgf("Response body: %s", func() string {
		body, _ := io.ReadAll(resp.Body)
		return string(body)
	}())

	if resp.StatusCode == http.StatusOK {
		return "Flags submitted successfully", nil
	} else {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error().Msgf("Unexpected response status: %d, body: %s", resp.StatusCode, body)
		return "", fmt.Errorf("unexpected response status: %d, body: %s", resp.StatusCode, body)
	}
}
