package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

type sessionPayload struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	SessionID    string    `json:"session_id"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Metadata     string    `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastActivity time.Time `json:"last_activity"`
}

func sessionPayloadFrom(v *domain.ChatSession) *sessionPayload {
	if v == nil {
		return nil
	}
	out := sessionPayload(*v)
	return &out
}

type statsPayload struct {
	TotalMessages       int64   `json:"total_messages"`
	TotalSessions       int64   `json:"total_sessions"`
	AvgTokensPerMessage float64 `json:"avg_tokens_per_message"`
	TotalTokensUsed     int64   `json:"total_tokens_used"`
}

func statsPayloadFrom(v *domain.UserStats) *statsPayload {
	if v == nil {
		return nil
	}
	out := statsPayload(*v)
	return &out
}
