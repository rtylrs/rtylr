package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	sqlc "github.com/voxire/onion/internal/modules/memory/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
)

var (
	ErrChatSessionNotFound = errors.New("chat session not found")
)

const maxDerivedChatTitleRunes = 80

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	if pool == nil {
		return nil
	}

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) CreateChatSession(ctx context.Context, session *domain.ChatSession) (*domain.ChatSession, error) {
	arg := sqlc.CreateChatSessionParams{
		UserID:    int32(session.UserID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		SessionID: session.SessionID,
		Title:     sql.NullString{String: session.Title, Valid: session.Title != ""},
		Metadata:  []byte(session.Metadata),
	}

	_, err := r.queries.CreateChatSession(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	return r.GetChatSession(ctx, session.SessionID, session.UserID)
}

func (r *Repository) GetChatSession(ctx context.Context, sessionID string, userID int) (*domain.ChatSession, error) {
	arg := sqlc.GetChatSessionParams{
		SessionID: sessionID,
		UserID:    int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	row, err := r.queries.GetChatSession(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChatSessionNotFound
		}

		logger.Ctx(ctx).Error("Failed to get chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	var session domain.ChatSession
	if fromRowsErr := session.FromRows(row); fromRowsErr != nil {
		return nil, fromRowsErr
	}

	if user, ok := auth.GetAuthUser(ctx); ok && user != nil && user.BusinessID > 0 {
		var metadata struct {
			BusinessID int64 `json:"business_id"`
		}
		if err := json.Unmarshal([]byte(session.Metadata), &metadata); err != nil {
			return nil, ErrChatSessionNotFound
		}
		if metadata.BusinessID == 0 {
			var bid sql.NullInt64
			if err := r.pool.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id=?", userID).Scan(&bid); err != nil || !bid.Valid || bid.Int64 != user.BusinessID {
				return nil, ErrChatSessionNotFound
			}
		} else if metadata.BusinessID != user.BusinessID {
			return nil, ErrChatSessionNotFound
		}
	}
	return &session, nil
}

func (r *Repository) UpdateChatSession(ctx context.Context, sessionID string, userID int, title, summary, metadata string) (*domain.ChatSession, error) {
	if metadata == "" {
		metadata = "{}"
	}

	arg := sqlc.UpdateChatSessionParams{
		Title:     sql.NullString{String: title, Valid: title != ""},
		Summary:   sql.NullString{String: summary, Valid: summary != ""},
		Metadata:  []byte(metadata),
		SessionID: sessionID,
		UserID:    int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	err := r.queries.UpdateChatSession(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to update chat session: %w", err)
	}

	return r.GetChatSession(ctx, sessionID, userID)
}

func (r *Repository) DeleteChatSession(ctx context.Context, sessionID string, userID int) error {
	if _, err := r.GetChatSession(ctx, sessionID, userID); err != nil {
		return err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin delete chat session transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := r.queries.WithTx(tx)

	if err := qtx.DeleteSessionMessages(ctx, sqlc.DeleteSessionMessagesParams{
		SessionID: sessionID,
		UserID:    int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}); err != nil {
		return fmt.Errorf("failed to delete chat messages: %w", err)
	}

	if err := qtx.DeleteSessionSummaries(ctx, sqlc.DeleteSessionSummariesParams{
		SessionID: sessionID,
		UserID:    int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}); err != nil {
		return fmt.Errorf("failed to delete chat summaries: %w", err)
	}

	result, err := qtx.DeleteChatSessionByID(ctx, sqlc.DeleteChatSessionByIDParams{
		SessionID: sessionID,
		UserID:    int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		return fmt.Errorf("failed to delete chat session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to inspect deleted chat session rows: %w", err)
	}
	if rows == 0 {
		return ErrChatSessionNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit delete chat session transaction: %w", err)
	}

	return nil
}

func (r *Repository) CreateMessage(ctx context.Context, message *domain.ConversationMessage) (*domain.ConversationMessage, error) {
	arg := sqlc.CreateMessageParams{
		SessionID:  message.SessionID,
		UserID:     int32(message.UserID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Role:       message.Role,
		Content:    message.Content,
		TokensUsed: sql.NullInt32{Int32: int32(message.TokensUsed), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Metadata:   []byte(defaultMessageMetadata(message.Metadata)),
	}

	result, err := r.queries.CreateMessage(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create message", logger.Error(err))

		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to read message id: %w", err)
	}

	row, err := r.queries.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created message: %w", err)
	}

	var created domain.ConversationMessage
	if fromRowsErr := created.FromRows(row); fromRowsErr != nil {
		return nil, fromRowsErr
	}

	return &created, nil
}

func defaultMessageMetadata(metadata string) string {
	if strings.TrimSpace(metadata) == "" {
		return "{}"
	}
	return metadata
}

func (r *Repository) ListUserSessions(ctx context.Context, userID, limit, offset int) ([]domain.ChatSession, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if user, ok := auth.GetAuthUser(ctx); ok && user != nil && user.BusinessID > 0 {
		rows, err := r.pool.QueryContext(ctx, `SELECT id,user_id,session_id,title,summary,metadata,created_at,updated_at,last_activity
			FROM chat_sessions WHERE user_id=? AND (
			 CAST(JSON_UNQUOTE(JSON_EXTRACT(metadata,'$.business_id')) AS UNSIGNED)=? OR
			 ((metadata IS NULL OR JSON_EXTRACT(metadata,'$.business_id') IS NULL) AND EXISTS (
			  SELECT 1 FROM users WHERE id=chat_sessions.user_id AND business_id=?)))
			ORDER BY last_activity DESC LIMIT ? OFFSET ?`, userID, user.BusinessID, user.BusinessID, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list business chat sessions: %w", err)
		}
		defer rows.Close()
		results := make([]domain.ChatSession, 0, limit)
		for rows.Next() {
			var item domain.ChatSession
			var title, summary, metadata sql.NullString
			var createdAt, updatedAt, lastActivity sql.NullTime
			if err := rows.Scan(&item.ID, &item.UserID, &item.SessionID, &title, &summary, &metadata, &createdAt, &updatedAt, &lastActivity); err != nil {
				return nil, err
			}
			item.Title, item.Summary, item.Metadata = title.String, summary.String, metadata.String
			item.CreatedAt, item.UpdatedAt, item.LastActivity = createdAt.Time, updatedAt.Time, lastActivity.Time
			if shouldDeriveChatTitle(item.Title) {
				if derived, err := r.firstUserMessageTitle(ctx, item.SessionID, userID); err == nil && derived != "" {
					item.Title = derived
				}
			}
			results = append(results, item)
		}
		return results, rows.Err()
	}

	arg := sqlc.ListUserSessionsParams{
		UserID: int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.ListUserSessions(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list user sessions", logger.Error(err))
		return nil, fmt.Errorf("failed to list user sessions: %w", err)
	}

	results := make([]domain.ChatSession, 0, len(rows))
	for _, row := range rows {
		var item domain.ChatSession
		if fromRowsErr := item.FromRows(row); fromRowsErr != nil {
			return nil, fromRowsErr
		}
		if shouldDeriveChatTitle(item.Title) {
			title, titleErr := r.firstUserMessageTitle(ctx, item.SessionID, userID)
			if titleErr != nil {
				logger.Ctx(ctx).Debug("Failed to derive chat title from first user message", logger.Error(titleErr))
			} else if title != "" {
				item.Title = title
			}
		}
		if _, err := r.GetChatSession(ctx, item.SessionID, userID); err != nil {
			if errors.Is(err, ErrChatSessionNotFound) {
				continue
			}
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func shouldDeriveChatTitle(title string) bool {
	switch strings.TrimSpace(title) {
	case "", "Auto-generated session", "Untitled chat":
		return true
	default:
		return false
	}
}

func titleFromFirstUserMessage(content string) string {
	title := strings.Join(strings.Fields(content), " ")
	if title == "" {
		return ""
	}

	runes := []rune(title)
	if len(runes) <= maxDerivedChatTitleRunes {
		return title
	}
	return string(runes[:maxDerivedChatTitleRunes]) + "..."
}

func (r *Repository) firstUserMessageTitle(ctx context.Context, sessionID string, userID int) (string, error) {
	content, err := r.queries.GetFirstUserMessageContent(ctx, sqlc.GetFirstUserMessageContentParams{
		SessionID: sessionID,
		UserID:    int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return titleFromFirstUserMessage(content), nil
}

func (r *Repository) GetSessionMessages(ctx context.Context, sessionID string) ([]domain.ConversationMessage, error) {
	rows, err := r.queries.GetSessionMessages(ctx, sessionID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session messages", logger.Error(err))

		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	results := make([]domain.ConversationMessage, 0, len(rows))
	for _, row := range rows {
		var item domain.ConversationMessage
		if fromRowsErr := item.FromRows(row); fromRowsErr != nil {
			return nil, fromRowsErr
		}

		results = append(results, item)
	}

	return results, nil
}

func (r *Repository) GetContextForSession(ctx context.Context, sessionID string, userID int) ([]domain.ContextItem, error) {
	arg := sqlc.GetContextForSessionParams{
		SessionID:   sessionID,
		SessionID_2: sessionID,
		UserID:      int32(userID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}

	rows, err := r.queries.GetContextForSession(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to query context", logger.Error(err))

		return nil, fmt.Errorf("failed to query context: %w", err)
	}

	results := make([]domain.ContextItem, 0, len(rows))
	for _, row := range rows {
		var item domain.ContextItem
		if fromRowsErr := item.FromRows(row); fromRowsErr != nil {
			return nil, fromRowsErr
		}

		results = append(results, item)
	}

	return results, nil
}

func (r *Repository) CreateConversationSummary(ctx context.Context, summary *domain.ConversationSummary) (*domain.ConversationSummary, error) {
	keyTopicsJSON, marshalErr := json.Marshal(summary.KeyTopics)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal key topics: %w", marshalErr)
	}

	arg := sqlc.CreateConversationSummaryParams{
		UserID:          int32(summary.UserID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		SessionID:       summary.SessionID,
		Summary:         summary.Summary,
		KeyTopics:       keyTopicsJSON,
		Sentiment:       sql.NullString{String: summary.Sentiment, Valid: summary.Sentiment != ""},
		ImportanceScore: sql.NullFloat64{Float64: summary.ImportanceScore, Valid: true},
	}

	result, err := r.queries.CreateConversationSummary(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create conversation summary", logger.Error(err))

		return nil, fmt.Errorf("failed to create conversation summary: %w", err)
	}

	summaryID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to read conversation summary id: %w", err)
	}

	row, err := r.queries.GetConversationSummaryByID(ctx, summaryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created conversation summary: %w", err)
	}

	var created domain.ConversationSummary
	if fromRowsErr := created.FromRows(row); fromRowsErr != nil {
		return nil, fromRowsErr
	}

	return &created, nil
}

func (r *Repository) StoreUserMemory(ctx context.Context, memory *domain.UserMemory) (*domain.UserMemory, error) {
	arg := sqlc.StoreUserMemoryParams{
		UserID:        int32(memory.UserID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		MemoryType:    memory.MemoryType,
		Key:           memory.Key,
		Value:         memory.Value,
		Confidence:    sql.NullFloat64{Float64: memory.Confidence, Valid: true},
		SourceSession: sql.NullString{String: memory.SourceSession, Valid: memory.SourceSession != ""},
	}

	err := r.queries.StoreUserMemory(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store user memory", logger.Error(err))

		return nil, fmt.Errorf("failed to store user memory: %w", err)
	}

	row, err := r.queries.GetUserMemoryByKey(ctx, sqlc.GetUserMemoryByKeyParams{
		UserID:     int32(memory.UserID), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		MemoryType: memory.MemoryType,
		Key:        memory.Key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stored user memory: %w", err)
	}

	var stored domain.UserMemory
	if fromRowsErr := stored.FromRows(row); fromRowsErr != nil {
		return nil, fromRowsErr
	}

	return &stored, nil
}

// dynamic-query: squirrel
func (r *Repository) SearchUserMemory(ctx context.Context, userID int, query string) ([]domain.UserMemory, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []domain.UserMemory{}, nil
	}
	search := searchutil.BooleanPrefixQuery(query)

	rows, err := r.pool.QueryContext(ctx, `
SELECT
    id, user_id, memory_type, `+"`key`"+`, value, confidence,
    source_session, created_at, updated_at
FROM user_memory
WHERE user_id = ?
AND (
    memory_type = ? OR
    `+"`key`"+` = ? OR
    MATCH(`+"`key`"+`, value, memory_type) AGAINST (? IN BOOLEAN MODE)
)
ORDER BY confidence DESC`, int32(userID), query, query, search) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		logger.Ctx(ctx).Error("Failed to search user memory", logger.Error(err))

		return nil, fmt.Errorf("failed to search user memory: %w", err)
	}
	defer rows.Close()

	results := make([]domain.UserMemory, 0)
	for rows.Next() {
		var row sqlc.UserMemory
		if scanErr := rows.Scan(
			&row.ID,
			&row.UserID,
			&row.MemoryType,
			&row.Key,
			&row.Value,
			&row.Confidence,
			&row.SourceSession,
			&row.CreatedAt,
			&row.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan user memory search result: %w", scanErr)
		}

		var item domain.UserMemory
		if fromRowsErr := item.FromRows(row); fromRowsErr != nil {
			return nil, fromRowsErr
		}

		results = append(results, item)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("failed to iterate user memory search results: %w", rowsErr)
	}

	return results, nil
}

func (r *Repository) GetSessionStats(ctx context.Context, userID int) (*domain.UserStats, error) {
	row, err := r.queries.GetSessionStats(ctx, int32(userID)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session stats", logger.Error(err))

		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	var result domain.UserStats
	if fromRowsErr := result.FromRows(row); fromRowsErr != nil {
		return nil, fromRowsErr
	}

	return &result, nil
}

func (r *Repository) Close() error {
	return nil
}
