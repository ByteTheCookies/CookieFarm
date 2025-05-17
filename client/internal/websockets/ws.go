package websockets

import (
	"net/http"
	"time"

	"github.com/ByteTheCookies/cookieclient/internal/config"
	"github.com/ByteTheCookies/cookieclient/internal/logger"
	"github.com/gorilla/websocket"
)

const (
	FlagEvent = "flag"
)

func GetConnection() (*websocket.Conn, error) {
	const maxRetries = 3
	var conn *websocket.Conn
	var err error

	for attempts := 0; attempts < maxRetries; attempts++ {
		conn, _, err = websocket.DefaultDialer.Dial("ws://"+*config.HostServer+"/ws", http.Header{
			"Cookie": []string{"token=" + config.Token},
		})
		if err == nil {
			return conn, nil
		}

		logger.Log.Warn().Err(err).Int("attempt", attempts+1).Int("maxRetries", maxRetries).Msg("Error connecting to WebSocket, retrying...")
		time.Sleep(time.Second * time.Duration(attempts+1)) // Exponential backoff
	}

	logger.Log.Error().Err(err).Msg("Failed to connect to WebSocket after multiple attempts")
	return nil, err
}

// func ConnectToWebSocket() error {
// 	conn, _, err := websocket.DefaultDialer.Dial("ws://"+*config.HostServer+"/ws", nil)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()

// 	go func() {
// 		for {
// 			_, msg, err := conn.ReadMessage()
// 			if err != nil {
// 				logger.Log.Error().Err(err).Msg("Error reading message")
// 				break
// 			}
// 			fmt.Printf("Received: %s\n", msg)
// 		}
// 	}()

// 	go func() {
// 		for {
// 		}
// 	}()

// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)

// 	<-interrupt
// 	logger.Log.Info().Msg("Interruption signal received, closing connection...")
// 	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

// 	return nil
// }
