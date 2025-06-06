// Package flagparser provides functions to parse flags from JSON output.
package flagparser

import (
	"errors"
	"fmt"
	"time"

	"github.com/ByteTheCookies/cookieclient/internal/api"
	"github.com/ByteTheCookies/cookieclient/internal/config"
	json "github.com/bytedance/sonic"
)

// ParsedFlagOutput represents the output of a parsed flag returned
// by an exploit run in the exploit_manager, ready to be submitted.
type ParsedFlagOutput struct {
	TeamID      uint16 `json:"team_id"`      // ID of the team the flag was extracted from
	PortService uint16 `json:"port_service"` // Port of the service that produced the flag
	Status      string `json:"status"`       // Status of the flag submission (eg "success", "failed", "error", "fatal")
	FlagCode    string `json:"flag_code"`    // The actual flag string
	NameService string `json:"name_service"` // Human-readable name of the service
	Message     string `json:"message"`      // Additional message or error information
}

// ParseLine parses a JSON line into a Flag struct.
func ParseLine(line string) (api.Flag, string, error) {
	var out ParsedFlagOutput
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		return api.Flag{}, "invalid", fmt.Errorf("invalid JSON format: %w", err)
	}

	switch out.Status {
	case "info":
		return api.Flag{}, out.Status, errors.New(out.Message)
	case "failed":
		return api.Flag{}, out.Status, fmt.Errorf("flag submission failed for team %d on the %s: %s",
			out.TeamID, config.MapPortToService(out.PortService), out.Message)
	case "error":
		return api.Flag{}, out.Status, fmt.Errorf("flag submission error: %s", out.Message)
	case "fatal":
		return api.Flag{}, out.Status, fmt.Errorf("fatal error in the exploiter: %s", out.Message)
	case "success":
		return api.Flag{
			FlagCode:     out.FlagCode,
			ServiceName:  out.NameService,
			PortService:  out.PortService,
			SubmitTime:   uint64(time.Now().Unix()),
			ResponseTime: 0,
			Status:       "UNSUBMITTED",
			TeamID:       out.TeamID,
		}, out.Status, nil
	default:
		return api.Flag{}, "unknown", fmt.Errorf("unhandled status: %s", out.Status)
	}
}
