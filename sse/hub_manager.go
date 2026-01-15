package sse

import (
	"context"
	"sync"
	"time"
)

type hubManager struct {
	broker Broker
	opts   Options

	hubs   map[int64]*Hub
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func newHubManager(ctx context.Context, broker Broker, options Options) *hubManager {
	if ctx == nil {
		ctx = context.Background()
	}
	mctx, cancel := context.WithCancel(ctx)
	m := &hubManager{
		broker: broker,
		opts:   options,
		hubs:   make(map[int64]*Hub),
		ctx:    mctx,
		cancel: cancel,
	}

	go m.reaper()
	return m
}
func (hm *hubManager) getOrCreateHub(scopeID int64, patterns []string) *Hub {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hub, exists := hm.hubs[scopeID]
	if !exists {
		hub = newHub(hm.ctx, hm.broker, hm.opts, scopeID, patterns)
		hm.hubs[scopeID] = hub
	}

	return hub
}

func (hm *hubManager) reaper() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			hm.stopAll()
			return
		case <-t.C:
			var idle []*Hub
			hm.mu.Lock()
			for id, hub := range hm.hubs {
				if hub.isIdle(hm.opts.HubIdleTimeout) {
					idle = append(idle, hub)
					delete(hm.hubs, id)
				}
			}
			hm.mu.Unlock()

			for _, hub := range idle {
				hub.stop()
			}
		}
	}
}

func (hm *hubManager) stopAll() {
	hm.cancel()

	var hubs []*Hub
	hm.mu.Lock()
	for id, hub := range hm.hubs {
		hubs = append(hubs, hub)
		delete(hm.hubs, id)
	}
	hm.mu.Unlock()

	for _, hub := range hubs {
		hub.stop()
	}
}
