package sse

import (
	"sync"
	"time"
)

type hubManager struct {
	broker Broker
	opts   Options

	hubs map[int64]*Hub
	mu   sync.RWMutex
}

func newHubManager(broker Broker, options Options) *hubManager {
	m := &hubManager{
		broker: broker,
		opts:   options,
		hubs:   make(map[int64]*Hub),
	}

	go m.reaper()
	return m
}
func (hm *hubManager) getOrCreateHub(scopeID int64, patterns []string) *Hub {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hub, exists := hm.hubs[scopeID]
	if !exists {
		hub = newHub(hm.broker, hm.opts, scopeID, patterns)
		hm.hubs[scopeID] = hub
	}

	return hub
}

func (hm *hubManager) reaper() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()

	for range t.C {
		hm.mu.Lock()
		for id, hub := range hm.hubs {
			if hub.isIdle(hm.opts.HubIdleTimeout) {
				hub.stop()
				delete(hm.hubs, id)
			}
		}
		hm.mu.Unlock()
	}
}
