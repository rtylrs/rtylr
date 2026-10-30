package domain

import (
	"encoding/json"
	"time"
)

type Action struct {
	ID, RunID, DecidedBy int64
	ToolName, Status     string
	Input, Result        json.RawMessage
	ExpiresAt, CreatedAt time.Time
	DecidedAt            *time.Time
}
type RunView struct {
	ID                                 int64
	TriggerKind, Status, Report, Error string
	ToolCalls                          int32
	Tokens                             int64
	CreatedAt                          time.Time
	StartedAt, FinishedAt              *time.Time
}
