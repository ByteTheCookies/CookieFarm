package cnats

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/core"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type NATSServer struct {
	Conn   *nats.Conn
	Server *server.Server
}

func NewNATSServer() *NATSServer {
	return &NATSServer{
		Conn:   nil,
		Server: nil,
	}
}

// RunEmbeddedServer starts an embedded NATS server with JetStream enabled.
// If inProcess is true, it will run the server in-process; otherwise, it will
func (ns *NATSServer) RunEmbeddedServer(inProcess bool, enableLogging bool) error {
	var err error

	secretToken := func(n int) string {
		bytes := make([]byte, n)
		if _, err := rand.Read(bytes); err != nil {
			panic("unable to generate secure random token: " + err.Error())
		}
		return hex.EncodeToString(bytes)
	}(32)

	opts := &server.Options{
		ServerName:      "nats_server",
		JetStream:       true,
		JetStreamDomain: "jetcookie",
		StoreDir:        "./nats-data",
		Authorization:   secretToken,
	}

	ns.Server, err = server.NewServer(opts)
	if err != nil {
		return err
	}

	if enableLogging {
		ns.Server.ConfigureLogger()
	}
	go ns.Server.Start()

	if !ns.Server.ReadyForConnections(5 * time.Second) {
		return err
	}

	clientOpts := []nats.Option{
		nats.Token(secretToken),
	}
	if inProcess {
		clientOpts = append(clientOpts, nats.InProcessServer(ns.Server))
	}

	ns.Conn, err = nats.Connect(nats.DefaultURL, clientOpts...)
	if err != nil {
		return err
	}

	config.NATSToken = secretToken

	return err
}

// ShutdownEmbeddedServer gracefully shuts down the embedded NATS server and closes the connection.
func (ns *NATSServer) ShutdownEmbeddedServer() {
	if ns.Conn != nil {
		ns.Conn.Close()
	}

	if ns.Server != nil {
		ns.Server.Shutdown()
	}
}

func SetupNATS() (*NATSServer, error) {
	logger.Log.Debug().Msg("Setting up NATS server")
	ns := NewNATSServer()
	if err := ns.RunEmbeddedServer(true, config.Debug); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start embedded NATS server")
	}

	js, err := ns.Conn.JetStream()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to create JetStream context")
	}

	js.AddStream(&nats.StreamConfig{
		Name:     "cookiefarm",
		Subjects: []string{"cookiefarm.>"},
	})

	go core.ListenForFlags(ns.Conn)

	return ns, nil
}
