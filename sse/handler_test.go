package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEHandlerSendsEvent(t *testing.T) {
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

	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	if _, err := readLineWithTimeout(reader, 1*time.Second); err != nil {
		t.Fatalf("failed to read retry line: %v", err)
	}
	if _, err := readLineWithTimeout(reader, 1*time.Second); err != nil {
		t.Fatalf("failed to read retry separator: %v", err)
	}

	payload, _ := json.Marshal(map[string]any{"id": 123})
	if err := server.Publisher().PublishEvent(context.Background(), "scope:1:students", Event{
		EventType: "students.changed",
		Data:      payload,
	}); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	line, err := readLineWithTimeout(reader, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to read event line: %v", err)
	}
	if line != "event: app.students.changed" {
		t.Fatalf("unexpected event line: %s", line)
	}

	line, err = readLineWithTimeout(reader, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to read data line: %v", err)
	}
	if line != "data: {\"id\":123}" {
		t.Fatalf("unexpected data line: %s", line)
	}
}

func readLineWithTimeout(reader *bufio.Reader, timeout time.Duration) (string, error) {
	lineCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		line, err := reader.ReadString('\n')
		if err != nil {
			errCh <- err
			return
		}
		lineCh <- strings.TrimSpace(line)
	}()

	select {
	case line := <-lineCh:
		return line, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", context.DeadlineExceeded
	}
}
