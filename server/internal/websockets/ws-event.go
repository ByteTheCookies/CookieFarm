package websockets

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ByteTheCookies/cookieserver/internal/database"
	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/ByteTheCookies/cookieserver/internal/models"
)

const (
	FlagMessage  = "flag"
	FlagResponse = "flag_response"
)

// FlagHandler will send out a message to all other participants in the chat
func FlagHandler(event Event, client *Client) error {
	var flag models.Flag
	if err := json.Unmarshal(event.Payload, &flag); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	if err := database.AddFlag(flag); err != nil {
		logger.Log.Error().Err(err).Msg("DB insert failed in flag handler")
		return err
	}

	var responseMessage NewMessageEvent

	responseMessage.Sent = time.Now()

	data, err := json.Marshal(responseMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = FlagResponse

	client.Egress <- data

	logger.Log.Info().Int("client", client.Number).Str("flag", flag.FlagCode).Uint16("team", flag.TeamID).Str("service name", flag.ServiceName).Uint16("port service", flag.PortService).Msg("Flag received and sent to DB")

	return nil
}
