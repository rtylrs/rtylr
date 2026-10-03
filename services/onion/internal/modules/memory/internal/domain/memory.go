package domain

import (
	"encoding/json"
	"strconv"
	"time"

	sqlc "github.com/voxire/onion/internal/modules/memory/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
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
	Metadata   string    `json:"metadata,omitempty"`
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	UserID     int       `json:"user_id"`
	Role       string    `json:"role"` // "user", "assistant", "system".
	Content    string    `json:"content"`
	TokensUsed int       `json:"tokens_used"`
	CreatedAt  time.Time `json:"created_at"`
}

func (c *ConversationMessage) FromRows(message sqlc.ConversationMessage) error {
	c.Metadata = string(message.Metadata)
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
	c.KeyTopics = convert.ToStringSlice(summary.KeyTopics)
	c.Sentiment = convert.ToString(summary.Sentiment)
	c.ImportanceScore = convert.ToFloat64(summary.ImportanceScore)
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
	u.Confidence = convert.ToFloat64(memory.Confidence)
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
	u.AvgTokensPerMessage = interfaceToFloat64(stats.AvgTokensPerMessage)
	u.TotalTokensUsed = interfaceToInt64(stats.TotalTokensUsed)

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

	c.Type = item.Type
	c.Role = item.Role
	c.Content = item.Content
	c.CreatedAt = createdAt
	c.Summary = item.Summary
	c.KeyTopics = parseJSONStringSlice(item.KeyTopics)
	c.MemoryType = item.MemoryType
	c.MemoryKey = item.MemoryKey
	c.MemoryValue = item.MemoryValue

	return nil
}

func parseJSONStringSlice(raw string) []string {
	if raw == "" || raw == "null" {
		return nil
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}

	return values
}

func interfaceToFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case []byte:
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			// The domain layer stays dependency-free (no logger), so an
			// unparseable database aggregate safely coerces to zero.
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			// See []byte case: default safely to zero to keep the domain pure.
			return 0
		}
		return f
	default:
		return 0
	}
}

func interfaceToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case nil:
		return 0
	case int64:
		return v
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			// The domain layer stays dependency-free (no logger), so an
			// unparseable database aggregate safely coerces to zero.
			return 0
		}
		return i
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// See []byte case: default safely to zero to keep the domain pure.
			return 0
		}
		return i
	default:
		return 0
	}
}
