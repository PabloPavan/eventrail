package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/PabloPavan/eventrail/sse"
)

type resolverFunc func(*http.Request) (*sse.Principal, error)

func (f resolverFunc) Resolve(r *http.Request) (*sse.Principal, error) {
	return f(r)
}

func parseScope(r *http.Request) (int64, error) {
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		return 1, nil
	}
	value, err := strconv.ParseInt(scope, 10, 64)
	if err != nil {
		return 0, errors.New("invalid scope")
	}
	return value, nil
}

func normalizeJSON(raw string) json.RawMessage {
	if raw == "" {
		return json.RawMessage(`{}`)
	}
	if json.Valid([]byte(raw)) {
		return json.RawMessage(raw)
	}
	return json.RawMessage(strconv.Quote(raw))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	broker := newInMemoryBroker()

	server, err := sse.NewServer(broker, sse.Options{
		Context: ctx,
		Resolver: resolverFunc(func(r *http.Request) (*sse.Principal, error) {
			scopeID, err := parseScope(r)
			if err != nil {
				return nil, err
			}
			return &sse.Principal{UserID: 1, ScopeID: scopeID}, nil
		}),
		Router: func(p *sse.Principal) []string {
			return []string{fmt.Sprintf("scope:%d:*", p.ScopeID)}
		},
		EventNamePrefix: "app",
	})
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/events", server.Handler())
	mux.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		scopeID, err := parseScope(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		eventType := r.URL.Query().Get("type")
		if eventType == "" {
			eventType = "demo.ping"
		}
		data := normalizeJSON(r.URL.Query().Get("data"))

		channel := fmt.Sprintf("scope:%d:demo", scopeID)
		pubCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := server.Publisher().PublishEvent(pubCtx, channel, sse.Event{
			EventType: eventType,
			Data:      data,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "SSE endpoint: /events?scope=1")
		fmt.Fprintln(w, "Publish: /publish?scope=1&type=demo.ping&data={\"id\":123}")
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		server.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Println("listening on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
