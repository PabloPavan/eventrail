package sse

import "encoding/json"

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
