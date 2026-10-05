package contextjson

import (
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

type Message struct {
	Metadata   string    `json:"metadata,omitempty"`
	ID         int       `json:"id"`
	SessionID  string    `json:"session_id"`
	UserID     int       `json:"user_id"`
	Role       string    `json:"role"` // "user", "assistant", "system".
	Content    string    `json:"content"`
	TokensUsed int       `json:"tokens_used"`
	CreatedAt  time.Time `json:"created_at"`
}

type Memory struct {
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

type Context struct {
	SessionID string    `json:"session_id"`
	UserID    int       `json:"user_id"`
	Messages  []Message `json:"messages"`
	Summary   string    `json:"summary"`
	Memories  []Memory  `json:"memories"`
}

func Memories(values []domain.UserMemory) []Memory {
	if values == nil {
		return nil
	}
	out := make([]Memory, len(values))
	for i, v := range values {
		out[i] = Memory(v)
	}
	return out
}

func FromDomain(v *domain.ConversationContext) *Context {
	if v == nil {
		return nil
	}
	out := &Context{SessionID: v.SessionID, UserID: v.UserID, Summary: v.Summary, Memories: Memories(v.Memories)}
	if v.Messages != nil {
		out.Messages = make([]Message, len(v.Messages))
		for i, m := range v.Messages {
			out.Messages[i] = Message(m)
		}
	}
	return out
}

func Encode(v *domain.ConversationContext) ([]byte, error) { return json.Marshal(FromDomain(v)) }
