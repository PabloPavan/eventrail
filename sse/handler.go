package sse

import (
	"fmt"
	"net/http"
	"time"
)

func newHandler(hubs *hubManager, opts Options) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := opts.Resolver.Resolve(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to resolve principal: %v", err), http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		for k, v := range opts.Headers {
			w.Header().Set(k, v)
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		hub := hubs.getOrCreateHub(principal.ScopeID, opts.Router(principal))
		client := hub.addClient(opts.ClientBufferSize)
		defer hub.removeClient(client)

		if opts.Hooks.OnClientConnect != nil {
			opts.Hooks.OnClientConnect(principal.ScopeID)
		}
		defer func() {
			if opts.Hooks.OnClientDisconnect != nil {
				opts.Hooks.OnClientDisconnect(principal.ScopeID)
			}
		}()

		_, _ = fmt.Fprintf(w, ": retry %d\n\n", opts.RetryMilliseconds)
		flusher.Flush()

		heartbeatTicker := time.NewTicker(opts.HeartbeatInterval)
		defer heartbeatTicker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-heartbeatTicker.C:
				_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()

			case msg, ok := <-client.messageChan:
				if !ok {
					return
				}
				eventType, data, err := opts.EventEncoder(msg)
				if err != nil {
					if opts.Hooks.OnError != nil {
						opts.Hooks.OnError(r.Context(), fmt.Errorf("failed to encode event: %w", err))
					}
					continue
				}
				writeSSE(w, eventType, data)
				flusher.Flush()
			}
		}
	})
}

func writeSSE(w http.ResponseWriter, eventType string, data []byte) {
	_, _ = fmt.Fprintf(w, "event: %s\n", eventType)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}
