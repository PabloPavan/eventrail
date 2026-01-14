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
	Resolver      PrincipalResolver
	ChannelRouter ChannelRouter

	HeartbeatInterval  time.Duration
	RetryeMilliseconds int
	Headers            map[string]string

	ClientBufferSize int
	Backpressure     BackpressurePolicy
	HubIdleTimeout   time.Duration

	EventEncoder EventEncoder

	Hooks Hooks
}

func DefaultOptions() *Options {
	return &Options{
		HeartbeatInterval:  30 * time.Second,
		RetryeMilliseconds: 2000,
		Headers:            map[string]string{},
		ClientBufferSize:   64,
		Backpressure:       BackpressureDisconnect,
		HubIdleTimeout:     5 * time.Minute,
	}
}
