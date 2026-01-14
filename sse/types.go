package sse

import "encoding/json"

type Event struct {
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data,omitempty"`
}
