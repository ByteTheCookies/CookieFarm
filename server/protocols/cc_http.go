//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ByteTheCookies/cookieserver/internal/models"
)

func Submit(host string, team_token string, flags []string) ([]models.ResponseProtocol, error) {
	jsonData, err := json.Marshal(flags)
	if err != nil {
		return nil, fmt.Errorf("error during marshalling: %w", err)
	}

	url := "http://" + host + "/flags"
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error during request creation: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Team-Token", team_token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during request submission: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error during response reading: %w", err)
	}

	var responses []models.ResponseProtocol
	// logger.Debug("Raw body %s", string(body))
	if err := json.Unmarshal(body, &responses); err != nil {
		return nil, fmt.Errorf("error during response parsing: %w", err)
	}

	for i := range responses {
		responses[i].Msg = strings.Split(responses[i].Msg, "]")[1]
	}

	return responses, nil
}
