package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	"github.com/voxire/onion/internal/platform/auth"
)

func projectIdentity(ctx context.Context) (int64, int64, error) {
	user, ok := auth.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 || user.BusinessID <= 0 {
		return 0, 0, errors.New("authenticated business identity is required")
	}
	return user.UserID, user.BusinessID, nil
}
func (r *Repository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	uid, bid, err := projectIdentity(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.QueryContext(ctx, "SELECT id,name,icon,color,context,updated_at FROM caret_projects WHERE user_id=? AND business_id=? AND archived=FALSE ORDER BY updated_at DESC,id", uid, bid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := []domain.Project{}
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Icon, &p.Color, &p.Context, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}
func (r *Repository) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	uid, bid, err := projectIdentity(ctx)
	if err != nil {
		return nil, err
	}
	var p domain.Project
	err = r.pool.QueryRowContext(ctx, "SELECT id,name,icon,color,context,updated_at FROM caret_projects WHERE id=? AND user_id=? AND business_id=? AND archived=FALSE", id, uid, bid).Scan(&p.ID, &p.Name, &p.Icon, &p.Color, &p.Context, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrChatSessionNotFound
	}
	return &p, err
}
func (r *Repository) SaveProject(ctx context.Context, p domain.Project, create bool) (*domain.Project, error) {
	uid, bid, err := projectIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if create {
		_, err = r.pool.ExecContext(ctx, "INSERT INTO caret_projects(id,user_id,business_id,name,icon,color,context) VALUES(?,?,?,?,?,?,?)", p.ID, uid, bid, p.Name, p.Icon, p.Color, p.Context)
	} else {
		if _, err = r.GetProject(ctx, p.ID); err != nil {
			return nil, err
		}
		_, err = r.pool.ExecContext(ctx, "UPDATE caret_projects SET name=?,icon=?,color=?,context=? WHERE id=? AND user_id=? AND business_id=? AND archived=FALSE", p.Name, p.Icon, p.Color, p.Context, p.ID, uid, bid)
	}
	if err != nil {
		return nil, err
	}
	return r.GetProject(ctx, p.ID)
}
func (r *Repository) ArchiveProject(ctx context.Context, id string) error {
	uid, bid, err := projectIdentity(ctx)
	if err != nil {
		return err
	}
	if _, err = r.GetProject(ctx, id); err != nil {
		return err
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `UPDATE chat_sessions SET metadata=JSON_REMOVE(metadata,'$.project_id'),updated_at=CURRENT_TIMESTAMP
		WHERE user_id=? AND CAST(JSON_UNQUOTE(JSON_EXTRACT(metadata,'$.business_id')) AS UNSIGNED)=?
		AND JSON_UNQUOTE(JSON_EXTRACT(metadata,'$.project_id'))=?`, uid, bid, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE caret_projects SET archived=TRUE WHERE id=? AND user_id=? AND business_id=? AND archived=FALSE", id, uid, bid); err != nil {
		return err
	}
	return tx.Commit()
}
func (r *Repository) SetChatProject(ctx context.Context, session string, userID int, projectID string) error {
	if _, err := r.GetChatSession(ctx, session, userID); err != nil {
		return err
	}
	if projectID != "" {
		if _, err := r.GetProject(ctx, projectID); err != nil {
			return err
		}
	}
	_, err := r.pool.ExecContext(ctx, `UPDATE chat_sessions SET metadata=JSON_SET(COALESCE(metadata,JSON_OBJECT()),'$.project_id',?),updated_at=CURRENT_TIMESTAMP WHERE session_id=? AND user_id=?`, projectID, session, userID)
	return err
}
func (r *Repository) AssistantProfile(ctx context.Context) (string, error) {
	uid, bid, err := projectIdentity(ctx)
	if err != nil {
		return "", err
	}
	var name, business, tz, currency string
	err = r.pool.QueryRowContext(ctx, "SELECT COALESCE(u.full_name,''),b.name,COALESCE(b.timezone,'UTC'),COALESCE(b.currency,'USD') FROM users u JOIN businesses b ON b.id=? WHERE u.id=?", bid, uid).Scan(&name, &business, &tz, &currency)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Authenticated person: %s\nCurrent business: %s\nBusiness timezone: %s\nBusiness currency: %s", name, business, tz, currency), nil
}

func (r *Repository) AssistantMemories(ctx context.Context, sessionID string) ([]domain.UserMemory, error) {
	user, ok := auth.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 || user.BusinessID <= 0 {
		return nil, errors.New("authenticated business identity is required")
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT m.id,m.user_id,m.memory_type,m.key,m.value,m.confidence,m.source_session,m.created_at,m.updated_at
		FROM user_memory m WHERE m.user_id=? AND m.confidence>0.7 AND (
		 m.memory_type IN ('preference','interest') OR EXISTS (
		  SELECT 1 FROM chat_sessions current_chat JOIN chat_sessions source_chat
		   ON source_chat.session_id=m.source_session AND source_chat.user_id=m.user_id
		  WHERE current_chat.session_id=? AND current_chat.user_id=? AND (
		   CAST(JSON_UNQUOTE(JSON_EXTRACT(current_chat.metadata,'$.business_id')) AS UNSIGNED)=CAST(JSON_UNQUOTE(JSON_EXTRACT(source_chat.metadata,'$.business_id')) AS UNSIGNED)
		   OR ((JSON_EXTRACT(current_chat.metadata,'$.business_id') IS NULL OR JSON_EXTRACT(source_chat.metadata,'$.business_id') IS NULL)
		    AND EXISTS (SELECT 1 FROM users u WHERE u.id=m.user_id AND u.business_id=?))))
		ORDER BY m.confidence DESC LIMIT 10`, user.UserID, sessionID, user.UserID, user.BusinessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	memories := make([]domain.UserMemory, 0, 10)
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
func (r *Repository) ProjectConversationContext(ctx context.Context, id, session string) (string, error) {
	p, err := r.GetProject(ctx, id)
	if err != nil {
		return "", err
	}
	uid, bid, err := projectIdentity(ctx)
	if err != nil {
		return "", err
	}
	rows, err := r.pool.QueryContext(ctx, `SELECT m.role,LEFT(m.content,2000) FROM conversation_messages m JOIN chat_sessions s ON s.session_id=m.session_id AND s.user_id=m.user_id WHERE s.user_id=? AND COALESCE(JSON_EXTRACT(s.metadata,'$.business_id'),0)=? AND JSON_UNQUOTE(JSON_EXTRACT(s.metadata,'$.project_id'))=? AND s.session_id<>? ORDER BY m.id DESC LIMIT 16`, uid, bid, id, session)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	type entry struct {
		Role    string
		Content string
	}
	entries := []entry{}
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.Role, &e.Content); err != nil {
			return "", err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	data, err := json.Marshal(struct {
		Name        string
		Notes       string
		RecentChats []entry
	}{p.Name, p.Context, entries})
	return string(data), err
}
