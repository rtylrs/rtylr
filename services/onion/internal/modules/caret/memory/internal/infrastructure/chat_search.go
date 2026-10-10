package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
)

// chatBusinessScope keeps a search inside the chats of the current business,
// the same rule the chat list uses. Parameters: business, legacy business,
// business.
const chatBusinessScope = `(CAST(JSON_UNQUOTE(JSON_EXTRACT(s.metadata,'$.business_id')) AS UNSIGNED)=? OR
	((s.metadata IS NULL OR JSON_EXTRACT(s.metadata,'$.business_id') IS NULL) AND ?=?))`

// SearchChats finds the user's past messages about query in this business's
// chats, best matches first. Full-text search handles words; a substring scan
// catches what it skips (short words, partial terms).
func (r *Repository) SearchChats(ctx context.Context, userID int, query string, limit int) ([]domain.ChatSearchHit, error) {
	user, ok := auth.GetAuthUser(ctx)
	if !ok || user == nil || user.BusinessID <= 0 {
		return nil, errors.New("authenticated business identity is required")
	}
	legacyBusiness, err := r.legacyBusiness(ctx, int64(userID))
	if err != nil {
		return nil, err
	}
	scope := []any{user.BusinessID, legacyBusiness, user.BusinessID}
	rows, err := r.pool.QueryContext(ctx, `SELECT m.session_id, COALESCE(s.title,''), m.role, m.content, m.created_at
		FROM conversation_messages m JOIN chat_sessions s ON s.session_id=m.session_id AND s.user_id=m.user_id
		WHERE m.user_id=? AND MATCH(m.content) AGAINST (? IN NATURAL LANGUAGE MODE) AND `+chatBusinessScope+`
		ORDER BY MATCH(m.content) AGAINST (? IN NATURAL LANGUAGE MODE) DESC, m.created_at DESC LIMIT ?`,
		append(append([]any{userID, query}, scope...), query, limit)...)
	if err != nil {
		return nil, err
	}
	hits, err := scanChatHits(rows, query)
	if err != nil || len(hits) > 0 {
		return hits, err
	}
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(query) + "%"
	rows, err = r.pool.QueryContext(ctx, `SELECT m.session_id, COALESCE(s.title,''), m.role, m.content, m.created_at
		FROM conversation_messages m JOIN chat_sessions s ON s.session_id=m.session_id AND s.user_id=m.user_id
		WHERE m.user_id=? AND m.content LIKE ? AND `+chatBusinessScope+`
		ORDER BY m.created_at DESC LIMIT ?`, append([]any{userID, pattern}, append(scope, limit)...)...)
	if err != nil {
		return nil, err
	}
	return scanChatHits(rows, query)
}

func scanChatHits(rows *sql.Rows, query string) ([]domain.ChatSearchHit, error) {
	defer rows.Close()
	hits := []domain.ChatSearchHit{}
	for rows.Next() {
		var hit domain.ChatSearchHit
		var content string
		var created sql.NullTime
		if err := rows.Scan(&hit.SessionID, &hit.Title, &hit.Role, &content, &created); err != nil {
			return nil, err
		}
		hit.Snippet, hit.CreatedAt = matchSnippet(content, query, 320), created.Time
		hits = append(hits, hit)
	}
	return hits, rows.Err()
}

// matchSnippet cuts a window of text around the first query word it finds.
func matchSnippet(content, query string, width int) string {
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) <= width {
		return content
	}
	lower := strings.ToLower(content)
	at := -1
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if i := strings.Index(lower, word); i >= 0 && (at < 0 || i < at) {
			at = i
		}
	}
	start := 0
	if at > 0 {
		start = max(0, utf8.RuneCountInString(content[:at])-width/3)
	}
	end := min(len(runes), start+width)
	snippet := string(runes[start:end])
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet += "..."
	}
	return snippet
}

// SetChatMetadataNumber sets one numeric metadata key on a chat in place, so
// concurrent background passes don't overwrite each other's keys.
func (r *Repository) SetChatMetadataNumber(ctx context.Context, sessionID string, userID int, key string, value int) error {
	for _, ch := range key {
		if ch != '_' && (ch < 'a' || ch > 'z') {
			return errors.New("metadata key must be lowercase letters and underscores")
		}
	}
	_, err := r.pool.ExecContext(ctx, `UPDATE chat_sessions SET metadata=JSON_SET(COALESCE(metadata, JSON_OBJECT()), ?, ?) WHERE session_id=? AND user_id=?`,
		"$."+key, value, sessionID, userID)
	return err
}
