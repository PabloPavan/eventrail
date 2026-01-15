package sse

import (
	"context"
	"sync"
	"time"
)

type client struct {
	messageCh chan []byte
}

type Hub struct {
	scopeID  int64
	patterns []string

	broker Broker
	opts   Options
	ctx    context.Context

	mu         sync.RWMutex
	clients    map[*client]struct{}
	lastActive time.Time

	sub     Subscription
	cancel  context.CancelFunc
	running bool
}

func newHub(ctx context.Context, broker Broker, options Options, scopeID int64, patterns []string) *Hub {
	return &Hub{
		scopeID:    scopeID,
		patterns:   patterns,
		broker:     broker,
		opts:       options,
		ctx:        ctx,
		clients:    make(map[*client]struct{}),
		lastActive: time.Now(),
	}
}

func (h *Hub) addClient(buf int) *client {
	h.mu.Lock()
	defer h.mu.Unlock()

	c := &client{messageCh: make(chan []byte, buf)}

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
		close(c.messageCh)
		delete(h.clients, c)
		h.lastActive = time.Now()
	}
}

func (h *Hub) start() {
	ctx, cancel := context.WithCancel(h.ctx)
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
	defer func() {
		h.mu.Lock()
		if h.sub == sub {
			h.running = false
			h.sub = nil
			h.cancel = nil
		}
		h.mu.Unlock()
	}()

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
	var toRemove []*client
	n := 0

	h.mu.RLock()
	switch h.opts.Backpressure {
	case BackpressureDrop:
		for c := range h.clients {
			select {
			case c.messageCh <- payload:
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
			case c.messageCh <- payload:
				n++
			default:
				toRemove = append(toRemove, c)
			}
		}
	}
	h.mu.RUnlock()

	h.mu.Lock()
	for _, c := range toRemove {
		if _, exists := h.clients[c]; exists {
			close(c.messageCh)
			delete(h.clients, c)
		}
	}
	h.lastActive = time.Now()
	h.mu.Unlock()

	if h.opts.Hooks.OnEventBroadcast != nil {
		h.opts.Hooks.OnEventBroadcast(h.scopeID, n)
	}
}

func (h *Hub) stop() {
	h.mu.Lock()
	cancel := h.cancel
	sub := h.sub
	h.running = false
	h.cancel = nil
	h.sub = nil

	for c := range h.clients {
		close(c.messageCh)
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
	return len(h.clients) == 0 && time.Since(h.lastActive) > timeout
}
