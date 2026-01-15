package sse

import "context"

type BrokerMsg struct {
	Pattern string
	Channel string
	Payload []byte
}

type Subscription interface {
	Channel() <-chan BrokerMsg
	Close() error
}

type Broker interface {
	Subscribe(ctx context.Context, patterns ...string) (Subscription, error)
	Publish(ctx context.Context, channel string, payload []byte) error
}
