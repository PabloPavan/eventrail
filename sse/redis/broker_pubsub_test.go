package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestBrokerPubSubPublishSubscribe(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	broker := NewBrokerPubSub(rdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:*")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

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
		if msg.Pattern != "scope:*" {
			t.Fatalf("unexpected pattern: %s", msg.Pattern)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	_ = sub.Close()
}

func TestBrokerPubSubCloseClosesChannel(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	broker := NewBrokerPubSub(rdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:*")
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
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for channel close")
	}
}
