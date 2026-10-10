package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

// contextMessageLimit bounds one context read. Older turns are already folded
// into the chat's rolling summary by compaction.
const contextMessageLimit = 400

// GetContextMessages loads a chat's newest messages for prompt building,
// oldest first. Image attachments are stripped in SQL: the context only needs
// text and tool steps, and images can be megabytes each.
func (r *Repository) GetContextMessages(ctx context.Context, sessionID string) ([]domain.ConversationMessage, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT id, session_id, user_id, role, content, tokens_used,
		CAST(COALESCE(JSON_REMOVE(metadata, '$.images'), JSON_OBJECT()) AS CHAR) AS metadata, created_at
		FROM conversation_messages WHERE session_id = ? ORDER BY id DESC LIMIT ?`, sessionID, contextMessageLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]domain.ConversationMessage, 0, 64)
	for rows.Next() {
		var message domain.ConversationMessage
		var tokens sql.NullInt64
		var createdAt sql.NullTime
		if err := rows.Scan(&message.ID, &message.SessionID, &message.UserID, &message.Role, &message.Content, &tokens, &message.Metadata, &createdAt); err != nil {
			return nil, err
		}
		message.TokensUsed, message.CreatedAt = int(tokens.Int64), createdAt.Time
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}
