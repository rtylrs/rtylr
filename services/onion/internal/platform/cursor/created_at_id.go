package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalid = errors.New("invalid cursor")

type CreatedAtID struct {
	CreatedAt time.Time
	ID        int64
}

type createdAtIDPayload struct {
	CreatedAt string `json:"created_at"`
	ID        int64  `json:"id"`
}

func EncodeCreatedAtID(value CreatedAtID) string {
	payload, err := json.Marshal(createdAtIDPayload{CreatedAt: value.CreatedAt.UTC().Format(time.RFC3339Nano), ID: value.ID})
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(payload)
}

func DecodeCreatedAtID(token string) (CreatedAtID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return CreatedAtID{}, ErrInvalid
	}

	var payload createdAtIDPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.ID <= 0 {
		return CreatedAtID{}, ErrInvalid
	}

	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil {
		return CreatedAtID{}, ErrInvalid
	}

	return CreatedAtID{CreatedAt: createdAt, ID: payload.ID}, nil
}
