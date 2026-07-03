//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"models"
	"net/http"
	"strings"

	"protocols"
)

// Submit forwards flags to a ForcAD gameserver HTTP receiver (https://github.com/pomo-mondreganto/ForcAD.git)
//
// ForcAD exposes `PUT /flags` (see backend/services/http_receiver/views.py),
// authenticated via the `X-Team-Token` header, with a JSON list of flag
// strings as body (max 100 per request).
//
// On success the receiver answers with a JSON array of objects shaped like:
//
//	[{"msg": "[<flag>] <message>", "flag": "<flag>"}]
//
// Note there is NO explicit status field: the verdict has to be derived from
// the human-readable message. On request-level failures (bad token, game not
// started, body too large, rate limit, ...) the server answers with a JSON
// object `{"error": "..."}` or an HTML error page instead, with a non-2xx
// status code. In those cases we return an error so the whole batch stays
// unsubmitted (status = 0) and is retried on the next tick
func Submit(url string, teamToken string, flags []string) ([]protocols.ResponseProtocol, error) {
	jsonData, err := json.Marshal(flags)
	if err != nil {
		return nil, fmt.Errorf("error during marshalling: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error during request creation: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Team-Token", teamToken)

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

	// Anything other than 2xx is a request-level failure (invalid token,
	// game not started, 413 body too large, 429 rate limit, nginx HTML page,
	// ...). Surface it as an error so the batch is retried instead of trying
	// to parse a non-array body
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("checker returned status %d: %s", resp.StatusCode, errorMessage(body))
	}

	var responses []struct {
		Flag string `json:"flag"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &responses); err != nil {
		// 200 but not the expected array: most likely a top-level
		// `{"error": "..."}` object. Report it so the batch is retried
		return nil, fmt.Errorf("error during response parsing: %w (body: %s)", err, errorMessage(body))
	}

	responsesParsed := make([]protocols.ResponseProtocol, len(responses))
	for i := range responses {
		responsesParsed[i].Flag = responses[i].Flag
		responsesParsed[i].Msg = cleanMessage(responses[i].Msg, responses[i].Flag)
		responsesParsed[i].Status = statusFromMessage(responses[i].Msg)
	}

	return responsesParsed, nil
}

// errorMessage extracts a human-readable message from a non-array response
// body, falling back to a trimmed raw snippet (e.g. an nginx HTML page)
func errorMessage(body []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &e); err == nil && e.Error != "" {
		return e.Error
	}

	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 200 {
		snippet = snippet[:200] + "..."
	}
	return snippet
}

// cleanMessage strips the leading "[<flag>] " prefix that ForcAD prepends to
// every message, without assuming it is always present.
func cleanMessage(msg, flag string) string {
	msg = strings.TrimSpace(msg)
	if prefix := "[" + flag + "]"; strings.HasPrefix(msg, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(msg, prefix))
	}
	return msg
}

func statusFromMessage(msg string) int64 {
	m := strings.ToLower(msg)

	switch {
	case strings.Contains(m, "accepted"):
		return models.StatusAccepted
	case strings.Contains(m, "already stolen"),
		strings.Contains(m, "your own"),
		strings.Contains(m, "invalid"),
		strings.Contains(m, "too old"):
		return models.StatusDenied
	case strings.Contains(m, "service is down"),
		strings.Contains(m, "not available"):
		return models.StatusNotValid
	default:
		// Unknown verdict: mark as Error (terminal) to avoid an infinite
		// resubmission loop, but it will show up in the logs.
		return models.StatusError
	}
}
