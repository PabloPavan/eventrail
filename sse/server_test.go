package sse

import (
	"net/http"
	"testing"
	"time"
)

func TestNewServerValidation(t *testing.T) {
	broker := newTestBroker()

	if _, err := NewServer(nil, Options{}); err == nil {
		t.Fatal("expected error for nil broker")
	}

	if _, err := NewServer(broker, Options{Router: func(*Principal) []string { return nil }}); err == nil {
		t.Fatal("expected error for nil resolver")
	}

	if _, err := NewServer(broker, Options{Resolver: resolverFunc(func(*http.Request) (*Principal, error) { return nil, nil })}); err == nil {
		t.Fatal("expected error for nil router")
	}
}

func TestNewServerAppliesDefaults(t *testing.T) {
	broker := newTestBroker()

	server, err := NewServer(broker, Options{
		Resolver: resolverFunc(func(*http.Request) (*Principal, error) {
			return &Principal{UserID: 1, ScopeID: 1}, nil
		}),
		Router: func(*Principal) []string { return []string{"scope:1:*"} },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server.opts.HeartbeatInterval != 30*time.Second {
		t.Fatalf("unexpected heartbeat interval: %v", server.opts.HeartbeatInterval)
	}
	if server.opts.RetryMilliseconds != 3000 {
		t.Fatalf("unexpected retry milliseconds: %d", server.opts.RetryMilliseconds)
	}
	if server.opts.ClientBufferSize != 128 {
		t.Fatalf("unexpected client buffer size: %d", server.opts.ClientBufferSize)
	}
	if server.opts.HubIdleTimeout != 5*time.Minute {
		t.Fatalf("unexpected hub idle timeout: %v", server.opts.HubIdleTimeout)
	}
	if server.opts.Headers == nil {
		t.Fatal("expected headers to be initialized")
	}
	if server.opts.Backpressure != BackpressureDrop {
		t.Fatalf("unexpected backpressure policy: %v", server.opts.Backpressure)
	}
	if server.opts.EventEncoder == nil {
		t.Fatal("expected event encoder to be set")
	}
}

func TestEventNamePrefixApplied(t *testing.T) {
	broker := newTestBroker()

	server, err := NewServer(broker, Options{
		Resolver: resolverFunc(func(*http.Request) (*Principal, error) {
			return &Principal{UserID: 1, ScopeID: 1}, nil
		}),
		Router:          func(*Principal) []string { return []string{"scope:1:*"} },
		EventNamePrefix: "app",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eventType, _, err := server.opts.EventEncoder([]byte(`{"event_type":"students.changed","data":{"id":1}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "app.students.changed" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
}

type resolverFunc func(*http.Request) (*Principal, error)

func (f resolverFunc) Resolve(r *http.Request) (*Principal, error) {
	return f(r)
}
