package domain

import "time"

type Buddy struct {
	ID              int64
	BusinessID      int64
	RunAsUserID     int64
	Enabled         bool
	Instructions    string
	IntervalMinutes int
	DailyRunCap     int
	NextRunAt       *time.Time
	LastRunAt       *time.Time
}

type Run struct {
	ID          int64
	BuddyID     int64
	BusinessID  int64
	TriggerKind string
	Attempts    int
}

type RunResult struct {
	Report    string
	Error     string
	ToolCalls int
	Tokens    int64
}

const (
	TriggerSchedule = "schedule"
	// MaxRunAttempts bounds how often an interrupted run is retried.
	MaxRunAttempts = 3
)
