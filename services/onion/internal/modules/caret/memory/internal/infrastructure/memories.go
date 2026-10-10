package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
)

// memoryVisibility matches what Caret loads into its context: the user's
// personal preferences everywhere, and facts learned in a chat of the current
// business. Parameters: user, user, legacy business, business.
const memoryVisibility = `m.user_id=? AND m.confidence>0.7 AND (
	m.memory_type IN ('preference','interest') OR EXISTS (
		SELECT 1 FROM chat_sessions source_chat
		WHERE source_chat.session_id=m.source_session AND source_chat.user_id=?
		AND COALESCE(CAST(JSON_UNQUOTE(JSON_EXTRACT(source_chat.metadata,'$.business_id')) AS UNSIGNED),?)=?
	)
)`

func (r *Repository) memoryScope(ctx context.Context) ([]any, int64, error) {
	user, ok := auth.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 || user.BusinessID <= 0 {
		return nil, 0, errors.New("authenticated business identity is required")
	}
	legacyBusiness, err := r.legacyBusiness(ctx, user.UserID)
	if err != nil {
		return nil, 0, err
	}
	return []any{user.UserID, user.UserID, legacyBusiness, user.BusinessID}, user.UserID, nil
}

// ListMemories returns every memory Caret can use for this user here, newest
// first.
func (r *Repository) ListMemories(ctx context.Context) ([]domain.UserMemory, error) {
	scope, _, err := r.memoryScope(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT m.id,m.user_id,m.memory_type,m.key,m.value,m.confidence,m.source_session,m.created_at,m.updated_at
		FROM user_memory m WHERE `+memoryVisibility+` ORDER BY m.updated_at DESC, m.id DESC LIMIT 200`, scope...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemories(rows)
}

// SaveMemory updates a visible memory's text, or creates a personal
// preference when id is 0.
func (r *Repository) SaveMemory(ctx context.Context, id int, value string) (*domain.UserMemory, error) {
	scope, userID, err := r.memoryScope(ctx)
	if err != nil {
		return nil, err
	}
	if id == 0 {
		result, err := r.pool.ExecContext(ctx, `INSERT INTO user_memory (user_id,memory_type,`+"`key`"+`,value,confidence) VALUES (?,?,?,?,1)`,
			userID, "preference", "note_"+uuid.NewString(), value)
		if err != nil {
			return nil, err
		}
		inserted, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		id = int(inserted)
	} else {
		args := append([]any{value, id}, scope...)
		result, err := r.pool.ExecContext(ctx, `UPDATE user_memory m SET m.value=?, m.updated_at=CURRENT_TIMESTAMP WHERE m.id=? AND `+memoryVisibility, args...)
		if err != nil {
			return nil, err
		}
		if affected, err := result.RowsAffected(); err != nil {
			return nil, err
		} else if affected == 0 {
			// MySQL reports 0 for an unchanged value too; confirm it exists.
			if _, err := r.visibleMemory(ctx, id, scope); err != nil {
				return nil, err
			}
		}
	}
	return r.visibleMemory(ctx, id, scope)
}

// DeleteMemory removes a visible memory so Caret stops using it.
func (r *Repository) DeleteMemory(ctx context.Context, id int) error {
	scope, _, err := r.memoryScope(ctx)
	if err != nil {
		return err
	}
	if _, err := r.visibleMemory(ctx, id, scope); err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, `DELETE FROM user_memory WHERE id=? AND user_id=?`, id, scope[0])
	return err
}

func (r *Repository) visibleMemory(ctx context.Context, id int, scope []any) (*domain.UserMemory, error) {
	args := append([]any{id}, scope...)
	rows, err := r.pool.QueryContext(ctx, `SELECT m.id,m.user_id,m.memory_type,m.key,m.value,m.confidence,m.source_session,m.created_at,m.updated_at
		FROM user_memory m WHERE m.id=? AND `+memoryVisibility, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memories, err := scanMemories(rows)
	if err != nil {
		return nil, err
	}
	if len(memories) == 0 {
		return nil, domain.ErrMemoryNotFound
	}
	return &memories[0], nil
}

func scanMemories(rows *sql.Rows) ([]domain.UserMemory, error) {
	memories := make([]domain.UserMemory, 0, 16)
	for rows.Next() {
		var memory domain.UserMemory
		var confidence float64
		var source sql.NullString
		if err := rows.Scan(&memory.ID, &memory.UserID, &memory.MemoryType, &memory.Key, &memory.Value, &confidence, &source, &memory.CreatedAt, &memory.UpdatedAt); err != nil {
			return nil, err
		}
		memory.Confidence, memory.SourceSession = confidence, source.String
		memories = append(memories, memory)
	}
	return memories, rows.Err()
}
