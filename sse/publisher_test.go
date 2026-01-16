package sse

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestPublisherPublishEventValidation(t *testing.T) {
	broker := newTestBroker()
	pub := NewPublisher(broker)

	if err := pub.PublishEvent(context.Background(), "", Event{EventType: "students.changed"}); err == nil {
		t.Fatal("expected error for empty channel")
	}

	if err := pub.PublishEvent(context.Background(), "scope:1:students", Event{}); err == nil {
		t.Fatal("expected error for empty event type")
	}
}

func TestPublisherPublishEventPayload(t *testing.T) {
	broker := newTestBroker()
	pub := NewPublisher(broker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:1:*")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	payload := json.RawMessage(`{"id":123}`)
	if err := pub.PublishEvent(context.Background(), "scope:1:students", Event{
		EventType: "students.changed",
		Data:      payload,
	}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		var evt Event
		if err := json.Unmarshal(msg.Payload, &evt); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}
		if evt.EventType != "students.changed" {
			t.Fatalf("unexpected event type: %s", evt.EventType)
		}
		if string(evt.Data) != string(payload) {
			t.Fatalf("unexpected payload: %s", string(evt.Data))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestPublisherPublishType(t *testing.T) {
	broker := newTestBroker()
	pub := NewPublisher(broker)

	if err := pub.PublishType(context.Background(), "scope:1:students", "students.changed"); err != nil {
		t.Fatalf("publish type failed: %v", err)
	}
}
