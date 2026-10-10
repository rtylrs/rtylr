package infrastructure

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

// SaveReplyFeedback records a verdict on a reply. A thumbs up replaces a
// thumbs down on the same reply and the reverse; "clear" removes either.
// Changing a verdict makes it pending again so the next distillation sees it.
func (r *Repository) SaveReplyFeedback(ctx context.Context, userID int, feedback domain.ReplyFeedback) error {
	if feedback.Rating == "up" || feedback.Rating == "down" || feedback.Rating == "clear" {
		if _, err := r.pool.ExecContext(ctx, `DELETE FROM caret_reply_feedback WHERE user_id=? AND message_id=? AND rating IN ('up','down') AND rating<>?`,
			userID, feedback.MessageID, feedback.Rating); err != nil {
			return err
		}
		if feedback.Rating == "clear" {
			return nil
		}
	}
	_, err := r.pool.ExecContext(ctx, `INSERT INTO caret_reply_feedback (user_id,session_id,message_id,rating,reason,prompt_excerpt,reply_excerpt)
		VALUES (?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE reason=VALUES(reason), prompt_excerpt=VALUES(prompt_excerpt), reply_excerpt=VALUES(reply_excerpt), distilled_at=NULL`,
		userID, feedback.SessionID, feedback.MessageID, feedback.Rating, feedback.Reason, feedback.Prompt, feedback.Reply)
	return err
}

// PendingReplyFeedback lists verdicts not yet distilled into lessons, oldest
// first.
func (r *Repository) PendingReplyFeedback(ctx context.Context, userID, limit int) ([]domain.ReplyFeedback, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT id,session_id,message_id,rating,reason,prompt_excerpt,reply_excerpt,created_at
		FROM caret_reply_feedback WHERE user_id=? AND distilled_at IS NULL ORDER BY id LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ReplyFeedback{}
	for rows.Next() {
		var item domain.ReplyFeedback
		var created sql.NullTime
		if err := rows.Scan(&item.ID, &item.SessionID, &item.MessageID, &item.Rating, &item.Reason, &item.Prompt, &item.Reply, &created); err != nil {
			return nil, err
		}
		item.CreatedAt = created.Time
		out = append(out, item)
	}
	return out, rows.Err()
}

// ReplaceLessons swaps the user's lesson memories for a new set and marks the
// feedback they came from as distilled, in one transaction.
func (r *Repository) ReplaceLessons(ctx context.Context, userID int, lessons []string, feedbackIDs []int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a committed transaction ignores rollback
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_memory WHERE user_id=? AND memory_type='lesson'`, userID); err != nil {
		return err
	}
	for _, lesson := range lessons {
		sum := sha256.Sum256([]byte(strings.ToLower(lesson)))
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_memory (user_id,memory_type,`key`,value,confidence) VALUES (?,'lesson',?,?,1) ON DUPLICATE KEY UPDATE value=VALUES(value)",
			userID, "lesson_"+hex.EncodeToString(sum[:8]), lesson); err != nil {
			return err
		}
	}
	if len(feedbackIDs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(feedbackIDs)), ",")
		args := []any{userID}
		for _, id := range feedbackIDs {
			args = append(args, id)
		}
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`UPDATE caret_reply_feedback SET distilled_at=CURRENT_TIMESTAMP WHERE user_id=? AND id IN (%s)`, placeholders), args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// UserLessons returns the user's current lessons, oldest first.
func (r *Repository) UserLessons(ctx context.Context, userID int) ([]string, error) {
	rows, err := r.pool.QueryContext(ctx, `SELECT value FROM user_memory WHERE user_id=? AND memory_type='lesson' ORDER BY id LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, rows.Err()
}
