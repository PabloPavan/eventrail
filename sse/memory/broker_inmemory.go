package memory

import (
	"context"
	"errors"
	"path"
	"sync"

	"github.com/PabloPavan/eventrail/sse"
)

type BrokerInMemory struct {
	mu   sync.RWMutex
	subs map[*memSubscription]struct{}
}

type memSubscription struct {
	broker   *BrokerInMemory
	patterns []string
	ch       chan sse.BrokerMsg
	once     sync.Once
}

func NewBrokerInMemory() *BrokerInMemory {
	return &BrokerInMemory{subs: make(map[*memSubscription]struct{})}
}

func (b *BrokerInMemory) Subscribe(ctx context.Context, patterns ...string) (sse.Subscription, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	sub := &memSubscription{
		broker:   b,
		patterns: append([]string(nil), patterns...),
		ch:       make(chan sse.BrokerMsg, 128),
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

func (b *BrokerInMemory) Publish(ctx context.Context, channel string, payload []byte) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	msg := sse.BrokerMsg{
		Channel: channel,
		Payload: payload,
	}

	b.mu.RLock()
	subs := make([]*memSubscription, 0, len(b.subs))
	for sub := range b.subs {
		subs = append(subs, sub)
	}
	b.mu.RUnlock()

	for _, sub := range subs {
		if !sub.matches(channel) {
			continue
		}
		select {
		case sub.ch <- msg:
		default:
		}
	}

	return nil
}

func (s *memSubscription) Channel() <-chan sse.BrokerMsg {
	return s.ch
}

func (s *memSubscription) Close() error {
	s.once.Do(func() {
		s.broker.mu.Lock()
		delete(s.broker.subs, s)
		s.broker.mu.Unlock()
		close(s.ch)
	})
	return nil
}

func (s *memSubscription) matches(channel string) bool {
	for _, pattern := range s.patterns {
		ok, err := path.Match(pattern, channel)
		if err == nil && ok {
			return true
		}
		if err != nil && pattern == channel {
			return true
		}
	}
	return false
}
