package sse

import (
	"context"
	"sync"
	"time"
)

type client struct {
	channelMessages chan []byte
}

type Hub struct {
	scopeID  int64
	patterns []string

	broker Broker
	opts   Options

	mu         sync.RWMutex
	clients    map[*client]struct{}
	lastActive time.Time

	sub     Subscription
	cancel  context.CancelFunc
	running bool
}

func newHub(broker Broker, options Options, scopeID int64, patterns []string) *Hub {
	return &Hub{
		scopeID:    scopeID,
		patterns:   patterns,
		broker:     broker,
		opts:       options,
		clients:    make(map[*client]struct{}),
		lastActive: time.Now(),
	}
}

func (h *Hub) addClient(buf int) *client {
	h.mu.Lock()
	defer h.mu.Unlock()

	c := &client{channelMessages: make(chan []byte, buf)}

	h.clients[c] = struct{}{}
	h.lastActive = time.Now()

	if !h.running {
		h.start()
	}

	return c
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[c]; exists {
		close(c.channelMessages)
		delete(h.clients, c)
		h.lastActive = time.Now()
	}
}

func (h *Hub) start() {
	ctx, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	h.running = true

	sub, err := h.broker.Subscribe(ctx, h.patterns...)
	if err != nil {
		h.running = false
		h.cancel = nil
		if h.opts.Hooks.OnError != nil {
			h.opts.Hooks.OnError(ctx, err)
		}
		return
	}

	h.sub = sub

	if h.opts.Hooks.OnHubStarted != nil {
		h.opts.Hooks.OnHubStarted(h.scopeID, h.patterns)
	}

	go h.run(ctx, sub)

}

func (h *Hub) run(ctx context.Context, sub Subscription) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub.Channel():
			if !ok {
				return
			}
			h.broadcast(msg.Payload)
		}
	}
}

func (h *Hub) broadcast(payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	n := 0
	switch h.opts.Backpressure {
	case BackpressureDrop:
		for c := range h.clients {
			select {
			case c.channelMessages <- payload:
				n++
			default:
				if h.opts.Hooks.OnClientDropped != nil {
					h.opts.Hooks.OnClientDropped(h.scopeID, "backpressure drop")
				}
			}
		}
	case BackpressureDisconnect:
		for c := range h.clients {
			select {
			case c.channelMessages <- payload:
				n++
			default:
				h.mu.RUnlock()
				h.removeClient(c)
				h.mu.RLock()
			}
		}
	}
	if h.opts.Hooks.OnEventBroadcast != nil {
		h.opts.Hooks.OnEventBroadcast(h.scopeID, n)
	}

	h.lastActive = time.Now()
}

func (h *Hub) stop() {
	h.mu.Lock()
	cancel := h.cancel
	sub := h.sub
	h.running = false
	h.cancel = nil
	h.sub = nil

	for c := range h.clients {
		close(c.channelMessages)
		delete(h.clients, c)
	}
	h.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if sub != nil {
		_ = sub.Close()
	}

	if h.opts.Hooks.OnHubStopped != nil {
		h.opts.Hooks.OnHubStopped(h.scopeID)
	}
}

func (h *Hub) isIdle(timeout time.Duration) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return !h.running && len(h.clients) == 0 && time.Since(h.lastActive) > timeout
}
