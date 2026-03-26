package domain

import (
	"time"

	"github.com/voxire/onion/internal/platform/utils/convert"
	sqlc "github.com/voxire/onion/internal/platforms/rtylr/memory/internal/infrastructure/persistence/sqlc"
)

type ChatSession struct {
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

func (c *ChatSession) FromRows(session sqlc.ChatSession) error {
	c.ID = int(session.ID)
	c.UserID = int(session.UserID)
	c.SessionID = session.SessionID
	c.Title = convert.ToString(session.Title)
	c.Summary = convert.ToString(session.Summary)
	c.Metadata = string(session.Metadata)
	c.CreatedAt = session.CreatedAt.Time
	c.UpdatedAt = session.UpdatedAt.Time
	c.LastActivity = session.LastActivity.Time

	return nil
}

type ConversationMessage struct {
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	UserID     int       `json:"user_id"`
	Role       string    `json:"role"` // "user", "assistant", "system".
	Content    string    `json:"content"`
	TokensUsed int       `json:"tokens_used"`
	CreatedAt  time.Time `json:"created_at"`
}

func (c *ConversationMessage) FromRows(message sqlc.ConversationMessage) error {
	c.ID = int(message.ID)
	c.SessionID = message.SessionID
	c.UserID = int(message.UserID)
	c.Role = message.Role
	c.Content = message.Content
	c.TokensUsed = int(message.TokensUsed.Int32)
	c.CreatedAt = message.CreatedAt.Time

	return nil
}

type ConversationSummary struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	SessionID       string    `json:"session_id"`
	Summary         string    `json:"summary"`
	KeyTopics       []string  `json:"key_topics"`
	Sentiment       string    `json:"sentiment"`
	ImportanceScore float64   `json:"importance_score"`
	CreatedAt       time.Time `json:"created_at"`
}

func (c *ConversationSummary) FromRows(summary sqlc.ConversationSummary) error {
	c.ID = int(summary.ID)
	c.UserID = int(summary.UserID)
	c.SessionID = summary.SessionID
	c.Summary = summary.Summary
	c.KeyTopics = summary.KeyTopics
	c.Sentiment = convert.ToString(summary.Sentiment)
	c.ImportanceScore = summary.ImportanceScore.Float64
	c.CreatedAt = summary.CreatedAt.Time

	return nil
}

type UserMemory struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	MemoryType    string    `json:"memory_type"` // "preference", "fact", "interest", etc.
	Key           string    `json:"key"`
	Value         string    `json:"value"`
	Confidence    float64   `json:"confidence"`
	SourceSession string    `json:"source_session"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (u *UserMemory) FromRows(memory sqlc.UserMemory) error {
	u.ID = int(memory.ID)
	u.UserID = int(memory.UserID)
	u.MemoryType = memory.MemoryType
	u.Key = memory.Key
	u.Value = memory.Value
	u.Confidence = memory.Confidence.Float64
	u.SourceSession = convert.ToString(memory.SourceSession)
	u.CreatedAt = memory.CreatedAt.Time
	u.UpdatedAt = memory.UpdatedAt.Time

	return nil
}

type ConversationContext struct {
	SessionID string                `json:"session_id"`
	UserID    int                   `json:"user_id"`
	Messages  []ConversationMessage `json:"messages"`
	Summary   string                `json:"summary"`
	Memories  []UserMemory          `json:"memories"`
}

type UserStats struct {
	TotalMessages       int64   `json:"total_messages"`
	TotalSessions       int64   `json:"total_sessions"`
	AvgTokensPerMessage float64 `json:"avg_tokens_per_message"`
	TotalTokensUsed     int64   `json:"total_tokens_used"`
}

func (u *UserStats) FromRows(stats sqlc.GetSessionStatsRow) error {
	u.TotalMessages = stats.TotalMessages
	u.TotalSessions = stats.TotalSessions
	u.AvgTokensPerMessage = stats.AvgTokensPerMessage
	u.TotalTokensUsed = stats.TotalTokensUsed

	return nil
}

type ContextItem struct {
	Type        string     `json:"type"` // "message", "summary", "memory".
	Role        string     `json:"role,omitempty"`
	Content     string     `json:"content"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	Summary     string     `json:"summary,omitempty"`
	KeyTopics   []string   `json:"key_topics,omitempty"`
	MemoryType  string     `json:"memory_type,omitempty"`
	MemoryKey   string     `json:"memory_key,omitempty"`
	MemoryValue string     `json:"memory_value,omitempty"`
}

func (c *ContextItem) FromRows(item sqlc.GetContextForSessionRow) error {
	var createdAt *time.Time
	if item.CreatedAt.Valid {
		createdAt = &item.CreatedAt.Time
	}

	// Handle interface{} types with type assertions
	var summary string

	if item.Summary != nil {
		if s, ok := item.Summary.(string); ok {
			summary = s
		}
	}

	var keyTopics []string

	if item.KeyTopics != nil {
		if kt, ok := item.KeyTopics.([]string); ok {
			keyTopics = kt
		}
	}

	var memoryType string

	if item.MemoryType != nil {
		if mt, ok := item.MemoryType.(string); ok {
			memoryType = mt
		}
	}

	var memoryKey string

	if item.MemoryKey != nil {
		if mk, ok := item.MemoryKey.(string); ok {
			memoryKey = mk
		}
	}

	var memoryValue string

	if item.MemoryValue != nil {
		if mv, ok := item.MemoryValue.(string); ok {
			memoryValue = mv
		}
	}

	c.Type = item.Type
	c.Role = item.Role
	c.Content = item.Content
	c.CreatedAt = createdAt
	c.Summary = summary
	c.KeyTopics = keyTopics
	c.MemoryType = memoryType
	c.MemoryKey = memoryKey
	c.MemoryValue = memoryValue

	return nil
}
