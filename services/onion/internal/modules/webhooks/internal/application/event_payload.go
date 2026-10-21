package application

import (
	"encoding/json"
)

func eventPayloadBytes(payload any) []byte {
	switch value := payload.(type) {
	case []byte:
		return value
	case json.RawMessage:
		return []byte(value)
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return encoded
	}
}
