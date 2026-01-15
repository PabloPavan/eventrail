package sse

import (
	"context"
	"time"
)

type BackpressurePolicy int

const (
	BackpressureDrop BackpressurePolicy = iota
	BackpressureDisconnect
)

type ChannelRouter func(p *Principal) []string

type EventEncoder func(raw []byte) (eventtype string, data []byte, err error)

type Hooks struct {
	OnClientConnect    func(scopeID int64)
	OnClientDisconnect func(scopeID int64)
	OnClientDropped    func(scopeID int64, reason string)
	OnEventBroadcast   func(scopeID int64, clients int)
	OnHubStarted       func(scopeID int64, patterns []string)
	OnHubStopped       func(scopeID int64)
	OnError            func(ctx context.Context, err error)
}

type Options struct {
	Resolver PrincipalResolver
	Router   ChannelRouter
	Context  context.Context

	HeartbeatInterval time.Duration
	RetryMilliseconds int
	Headers           map[string]string

	ClientBufferSize int
	Backpressure     BackpressurePolicy
	HubIdleTimeout   time.Duration

	EventEncoder EventEncoder

	Hooks Hooks
}

func applyDefaultOptions(opts *Options) {
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.HeartbeatInterval == 0 {
		opts.HeartbeatInterval = 30 * time.Second
	}
	if opts.RetryMilliseconds == 0 {
		opts.RetryMilliseconds = 3000
	}
	if opts.ClientBufferSize == 0 {
		opts.ClientBufferSize = 128
	}
	if opts.HubIdleTimeout == 0 {
		opts.HubIdleTimeout = 5 * time.Minute
	}
	if opts.EventEncoder == nil {
		opts.EventEncoder = defaultEventEncoder
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	if opts.Backpressure != BackpressureDrop && opts.Backpressure != BackpressureDisconnect {
		opts.Backpressure = BackpressureDisconnect
	}
}
