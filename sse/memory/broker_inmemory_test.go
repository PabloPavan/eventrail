package memory

import (
	"context"
	"testing"
	"time"
)

func TestBrokerInMemoryPublishCanceledContext(t *testing.T) {
	broker := NewBrokerInMemory()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := broker.Publish(ctx, "scope:1:students", []byte("hello")); err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestBrokerInMemoryPublishMatchesPattern(t *testing.T) {
	broker := NewBrokerInMemory()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:1:*")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	payload := []byte("hello")
	if err := broker.Publish(context.Background(), "scope:1:students", payload); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		if msg.Channel != "scope:1:students" {
			t.Fatalf("unexpected channel: %s", msg.Channel)
		}
		if string(msg.Payload) != string(payload) {
			t.Fatalf("unexpected payload: %s", string(msg.Payload))
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for message")
	}
}

func TestBrokerInMemoryPublishNonMatchingPattern(t *testing.T) {
	broker := NewBrokerInMemory()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:1:*")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	if err := broker.Publish(context.Background(), "scope:2:students", []byte("hello")); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case msg := <-sub.Channel():
		t.Fatalf("unexpected message: %v", msg)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestBrokerInMemoryCloseClosesChannel(t *testing.T) {
	broker := NewBrokerInMemory()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:1:*")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	if err := sub.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	select {
	case _, ok := <-sub.Channel():
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for channel close")
	}
}
