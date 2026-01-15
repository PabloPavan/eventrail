package sse

import (
	"encoding/json"
	"strings"
)

func defaultEventEncoder(raw []byte) (eventtype string, data []byte, err error) {
	evtType := "message"
	data = raw

	var evt Event
	if err := json.Unmarshal(raw, &evt); err != nil {
		return evtType, data, nil
	}
	if evt.EventType != "" {
		evtType = evt.EventType
		if len(evt.Data) == 0 {
			data = []byte(`{}`)
		} else {
			data = []byte(evt.Data)
		}
	}
	return evtType, data, nil
}

func applyEventNamePrefix(encoder EventEncoder, prefix string) EventEncoder {
	if encoder == nil {
		return nil
	}
	prefix = strings.TrimSpace(prefix)
	prefix = strings.TrimSuffix(prefix, ".")
	if prefix == "" {
		return encoder
	}

	return func(raw []byte) (string, []byte, error) {
		eventType, data, err := encoder(raw)
		if err != nil || eventType == "" {
			return eventType, data, err
		}
		if strings.HasPrefix(eventType, prefix+".") {
			return eventType, data, err
		}
		return prefix + "." + eventType, data, err
	}
}
