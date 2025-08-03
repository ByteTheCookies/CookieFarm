package cnats

import (
	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	Conn *nats.Conn
	Subs []*nats.Subscription
}

func NewNATSClient() *NATSClient {
	return &NATSClient{
		Conn: nil,
		Subs: make([]*nats.Subscription, 0),
	}
}

func (nc *NATSClient) Connect(token string, opts ...nats.Option) error {
	var err error
	clientOpts := []nats.Option{
		nats.Token(token),
	}

	nc.Conn, err = nats.Connect(nats.DefaultURL, clientOpts...)
	if err != nil {
		return err
	}

	return nil
}

func (nc *NATSClient) Disconnect() {
	nc.Conn.Close()
}

func (nc *NATSClient) Publish(subject string, msg []byte) error {
	if nc.Conn == nil {
		return nats.ErrConnectionClosed
	}
	err := nc.Conn.Publish(subject, msg)
	if err != nil {
		return err
	}

	return nil
}

func (nc *NATSClient) Subscribe(subject string, cb nats.MsgHandler) error {
	if nc.Conn == nil {
		return nats.ErrConnectionClosed
	}

	sub, err := nc.Conn.Subscribe(subject, cb)
	if err != nil {
		return err
	}

	nc.Subs = append(nc.Subs, sub)

	return nil
}
