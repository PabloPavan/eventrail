package sse

import "testing"

func TestDefaultEventEncoderPlainText(t *testing.T) {
	raw := []byte("hello")

	eventType, data, err := defaultEventEncoder(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "message" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
	if string(data) != string(raw) {
		t.Fatalf("unexpected data: %s", string(data))
	}
}

func TestDefaultEventEncoderJSONPayload(t *testing.T) {
	raw := []byte(`{"event_type":"students.changed","data":{"id":123}}`)

	eventType, data, err := defaultEventEncoder(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "students.changed" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
	if string(data) != `{"id":123}` {
		t.Fatalf("unexpected data: %s", string(data))
	}
}

func TestDefaultEventEncoderJSONWithoutData(t *testing.T) {
	raw := []byte(`{"event_type":"students.changed"}`)

	eventType, data, err := defaultEventEncoder(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "students.changed" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
	if string(data) != `{}` {
		t.Fatalf("unexpected data: %s", string(data))
	}
}

func TestDefaultEventEncoderInvalidJSON(t *testing.T) {
	raw := []byte(`{`)

	eventType, data, err := defaultEventEncoder(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "message" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
	if string(data) != string(raw) {
		t.Fatalf("unexpected data: %s", string(data))
	}
}

func TestApplyEventNamePrefix(t *testing.T) {
	encoder := applyEventNamePrefix(defaultEventEncoder, " app. ")

	eventType, data, err := encoder([]byte(`{"event_type":"students.changed","data":{"id":1}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "app.students.changed" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
	if string(data) != `{"id":1}` {
		t.Fatalf("unexpected data: %s", string(data))
	}
}

func TestApplyEventNamePrefixAlreadyApplied(t *testing.T) {
	encoder := applyEventNamePrefix(defaultEventEncoder, "app")

	eventType, _, err := encoder([]byte(`{"event_type":"app.students.changed","data":{"id":1}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "app.students.changed" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
}

func TestApplyEventNamePrefixEmptyPrefix(t *testing.T) {
	encoder := applyEventNamePrefix(defaultEventEncoder, "   ")

	eventType, _, err := encoder([]byte(`{"event_type":"students.changed","data":{"id":1}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eventType != "students.changed" {
		t.Fatalf("unexpected event type: %s", eventType)
	}
}
