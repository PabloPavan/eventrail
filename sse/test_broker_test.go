package sse

import (
	"context"
	"errors"
	"sync"
)

type testBroker struct {
	mu   sync.RWMutex
	subs map[*testSubscription]struct{}
}

type testSubscription struct {
	broker *testBroker
	ch     chan BrokerMsg
	once   sync.Once
}

func newTestBroker() *testBroker {
	return &testBroker{subs: make(map[*testSubscription]struct{})}
}

func (b *testBroker) Subscribe(ctx context.Context, _ ...string) (Subscription, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	sub := &testSubscription{
		broker: b,
		ch:     make(chan BrokerMsg, 128),
	}

	b.mu.Lock()
	b.subs[sub] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		_ = sub.Close()
	}()

	return sub, nil
}

func (b *testBroker) Publish(ctx context.Context, channel string, payload []byte) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	msg := BrokerMsg{
		Channel: channel,
		Payload: payload,
	}

	b.mu.RLock()
	subs := make([]*testSubscription, 0, len(b.subs))
	for sub := range b.subs {
		subs = append(subs, sub)
	}
	b.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.ch <- msg:
		default:
		}
	}

	return nil
}

func (s *testSubscription) Channel() <-chan BrokerMsg {
	return s.ch
}

func (s *testSubscription) Close() error {
	s.once.Do(func() {
		s.broker.mu.Lock()
		delete(s.broker.subs, s)
		s.broker.mu.Unlock()
		close(s.ch)
	})
	return nil
}
