package domain

import "time"

type ChatSession struct {
	ID           int
	UserID       int
	SessionID    string
	Title        string
	Summary      string
	Metadata     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastActivity time.Time
}

type ConversationMessage struct {
	Metadata   string
	ID         int
	SessionID  string
	UserID     int
	Role       string // "user", "assistant", "system".
	Content    string
	TokensUsed int
	CreatedAt  time.Time
}

type ConversationSummary struct {
	ID              int
	UserID          int
	SessionID       string
	Summary         string
	KeyTopics       []string
	Sentiment       string
	ImportanceScore float64
	CreatedAt       time.Time
}

type UserMemory struct {
	ID            int
	UserID        int
	MemoryType    string // "preference", "fact", "interest", etc.
	Key           string
	Value         string
	Confidence    float64
	SourceSession string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ConversationContext struct {
	SessionID string
	UserID    int
	Messages  []ConversationMessage
	Summary   string
	Memories  []UserMemory
}

type UserStats struct {
	TotalMessages       int64
	TotalSessions       int64
	AvgTokensPerMessage float64
	TotalTokensUsed     int64
}

type ContextItem struct {
	Type        string // "message", "summary", "memory".
	Role        string
	Content     string
	CreatedAt   *time.Time
	Summary     string
	KeyTopics   []string
	MemoryType  string
	MemoryKey   string
	MemoryValue string
}

// ChatSearchHit is one past message that matched a chat search.
type ChatSearchHit struct {
	SessionID string
	Title     string
	Role      string
	Snippet   string
	CreatedAt time.Time
}
