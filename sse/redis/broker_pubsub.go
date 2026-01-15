package redis

import (
	"context"

	"github.com/PabloPavan/eventrail/sse"
	"github.com/redis/go-redis/v9"
)

type BrokerPubSub struct {
	redisClient *redis.Client
}

func NewBrokerPubSub(redisClient *redis.Client) *BrokerPubSub {
	return &BrokerPubSub{redisClient: redisClient}
}

type redisSubscription struct {
	pubsub *redis.PubSub
	ch     <-chan *redis.Message
	out    chan sse.BrokerMsg
	ctx    context.Context
	cancel context.CancelFunc
}

func (b *BrokerPubSub) Subscribe(ctx context.Context, patterns ...string) (sse.Subscription, error) {
	pubsub := b.redisClient.PSubscribe(ctx, patterns...)
	cctx, cancel := context.WithCancel(ctx)

	sub := &redisSubscription{
		pubsub: pubsub,
		ch:     pubsub.Channel(),
		out:    make(chan sse.BrokerMsg, 128),
		ctx:    cctx,
		cancel: cancel,
	}

	go sub.forwardMessages()

	return sub, nil
}

func (r *redisSubscription) forwardMessages() {
	defer close(r.out)
	for {
		select {
		case <-r.ctx.Done():
			return
		case msg, ok := <-r.ch:
			if !ok {
				return
			}
			r.out <- sse.BrokerMsg{
				Pattern: msg.Pattern,
				Channel: msg.Channel,
				Payload: []byte(msg.Payload),
			}
		}
	}
}

func (r *redisSubscription) Channel() <-chan sse.BrokerMsg { return r.out }

func (r *redisSubscription) Close() error {
	r.cancel()
	return r.pubsub.Close()
}

func (b *BrokerPubSub) Publish(ctx context.Context, channel string, payload []byte) error {
	return b.redisClient.Publish(ctx, channel, payload).Err()
}
