package sse

import (
	"context"
	"encoding/json"
	"errors"
)

type Publisher struct {
	broker Broker
}

func NewPublisher(broker Broker) *Publisher {
	return &Publisher{broker: broker}
}

func (p *Publisher) PublishEvent(ctx context.Context, channel string, event Event) error {
	if channel == "" {
		return errors.New("channel cannot be empty")
	}
	if event.EventType == "" {
		return errors.New("event type cannot be empty")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.broker.Publish(ctx, channel, payload)
}

func (p *Publisher) PublishType(ctx context.Context, channel string, eventType string) error {
	return p.PublishEvent(ctx, channel, Event{EventType: eventType})
}
