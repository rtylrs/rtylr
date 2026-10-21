package infrastructure

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/caret/memory/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/utils/convert"
)

func chatSessionFromRow(row sqlc.ChatSession) domain.ChatSession {
	return domain.ChatSession{
		ID:           int(row.ID),
		UserID:       int(row.UserID),
		SessionID:    row.SessionID,
		Title:        convert.ToString(row.Title),
		Summary:      convert.ToString(row.Summary),
		Metadata:     string(row.Metadata),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		LastActivity: row.LastActivity.Time,
	}
}

func conversationMessageFromRow(row sqlc.ConversationMessage) domain.ConversationMessage {
	return domain.ConversationMessage{
		Metadata:   string(row.Metadata),
		ID:         int(row.ID),
		SessionID:  row.SessionID,
		UserID:     int(row.UserID),
		Role:       row.Role,
		Content:    row.Content,
		TokensUsed: int(row.TokensUsed.Int32),
		CreatedAt:  row.CreatedAt.Time,
	}
}

func conversationSummaryFromRow(row sqlc.ConversationSummary) domain.ConversationSummary {
	return domain.ConversationSummary{
		ID:              int(row.ID),
		UserID:          int(row.UserID),
		SessionID:       row.SessionID,
		Summary:         row.Summary,
		KeyTopics:       convert.ToStringSlice(row.KeyTopics),
		Sentiment:       convert.ToString(row.Sentiment),
		ImportanceScore: convert.ToFloat64(row.ImportanceScore),
		CreatedAt:       row.CreatedAt.Time,
	}
}

func userMemoryFromRow(row sqlc.UserMemory) domain.UserMemory {
	return domain.UserMemory{
		ID:            int(row.ID),
		UserID:        int(row.UserID),
		MemoryType:    row.MemoryType,
		Key:           row.Key,
		Value:         row.Value,
		Confidence:    convert.ToFloat64(row.Confidence),
		SourceSession: convert.ToString(row.SourceSession),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func userStatsFromRow(row sqlc.GetSessionStatsRow) domain.UserStats {
	return domain.UserStats{
		TotalMessages:       row.TotalMessages,
		TotalSessions:       row.TotalSessions,
		AvgTokensPerMessage: databaseFloat64(row.AvgTokensPerMessage),
		TotalTokensUsed:     databaseInt64(row.TotalTokensUsed),
	}
}

func contextItemFromRow(row sqlc.GetContextForSessionRow) domain.ContextItem {
	var createdAt *time.Time
	if row.CreatedAt.Valid {
		createdAt = &row.CreatedAt.Time
	}

	return domain.ContextItem{
		Type:        row.Type,
		Role:        row.Role,
		Content:     row.Content,
		CreatedAt:   createdAt,
		Summary:     row.Summary,
		KeyTopics:   parseJSONStringSlice(row.KeyTopics),
		MemoryType:  row.MemoryType,
		MemoryKey:   row.MemoryKey,
		MemoryValue: row.MemoryValue,
	}
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

func databaseFloat64(value interface{}) float64 {
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
		parsed, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return 0
		}
		return parsed
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func databaseInt64(value interface{}) int64 {
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
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}
