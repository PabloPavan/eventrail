package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestBrokerPubSubIntegrationRedis(t *testing.T) {
	addr := os.Getenv("INTEGRATION_REDIS_ADDR")
	if addr == "" {
		t.Skip("INTEGRATION_REDIS_ADDR not set")
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := waitForRedis(context.Background(), rdb, 2*time.Second); err != nil {
		t.Fatalf("redis not ready: %v", err)
	}

	broker := NewBrokerPubSub(rdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := broker.Subscribe(ctx, "scope:*")
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
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	_ = sub.Close()
}

func waitForRedis(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := client.Ping(ctx).Err(); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return context.DeadlineExceeded
		}
		time.Sleep(50 * time.Millisecond)
	}
}
