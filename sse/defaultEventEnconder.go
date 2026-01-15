package sse

import "encoding/json"

func DefaultEventEncoder(raw []byte) (eventtype string, data []byte, err error) {
	evtType := "message"
	data = raw

	var evt Event
	if err := json.Unmarshal(raw, &evt); err == nil && evt.EventType != "" {
		evtType = evt.EventType
		if len(evt.Data) == 0 {
			data = []byte(`{}`)
		} else {
			data = []byte(evt.Data)
		}
	} else {
		data = raw
	}
	return evtType, data, err
}
