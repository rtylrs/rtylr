package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	db "github.com/voxire/onion/internal/platform/database"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platforms/rtylr/memory/internal/domain"
	sqlc "github.com/voxire/onion/internal/platforms/rtylr/memory/internal/infrastructure/persistence/sqlc"
)

var (
	ErrDatabasePoolNotAvailable = errors.New("database pool not available")
	ErrChatSessionNotFound      = errors.New("chat session not found")
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
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
		UserID:    int32(session.UserID),
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
		UserID:    int32(userID),
	}

	result, err := db.NewReadExecutor[domain.ChatSession, sqlc.ChatSession](db.MainDatabase).Fetch(ctx, r.queries.GetChatSession, arg)
	if err != nil {
		if errors.Is(err, db.ErrNoResults) {
			return nil, ErrChatSessionNotFound
		}

		logger.Ctx(ctx).Error("Failed to get chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	return &result, nil
}

func (r *Repository) UpdateChatSession(ctx context.Context, sessionID string, userID int, title, summary, metadata string) (*domain.ChatSession, error) {
	arg := sqlc.UpdateChatSessionParams{
		Title:     sql.NullString{String: title, Valid: title != ""},
		Summary:   sql.NullString{String: summary, Valid: summary != ""},
		Metadata:  []byte(metadata),
		SessionID: sessionID,
		UserID:    int32(userID),
	}

	err := r.queries.UpdateChatSession(ctx, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to update chat session: %w", err)
	}

	return r.GetChatSession(ctx, sessionID, userID)
}

func (r *Repository) CreateMessage(ctx context.Context, message *domain.ConversationMessage) (*domain.ConversationMessage, error) {
	arg := sqlc.CreateMessageParams{
		SessionID:  message.SessionID,
		UserID:     int32(message.UserID),
		Role:       message.Role,
		Content:    message.Content,
		TokensUsed: sql.NullInt32{Int32: int32(message.TokensUsed), Valid: true},
		Metadata:   []byte("{}"),
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

func (r *Repository) GetSessionMessages(ctx context.Context, sessionID string) ([]domain.ConversationMessage, error) {
	results, err := db.NewReadExecutor[domain.ConversationMessage, sqlc.ConversationMessage](db.MainDatabase).FetchOrdered(ctx, r.queries.GetSessionMessages, sessionID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session messages", logger.Error(err))

		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	return results, nil
}

func (r *Repository) GetContextForSession(ctx context.Context, sessionID string, userID int) ([]domain.ContextItem, error) {
	arg := sqlc.GetContextForSessionParams{
		SessionID:   sessionID,
		SessionID_2: sessionID,
		UserID:      int32(userID),
	}

	results, err := db.NewReadExecutor[domain.ContextItem, sqlc.GetContextForSessionRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetContextForSession, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to query context", logger.Error(err))

		return nil, fmt.Errorf("failed to query context: %w", err)
	}

	return results, nil
}

func (r *Repository) CreateConversationSummary(ctx context.Context, summary *domain.ConversationSummary) (*domain.ConversationSummary, error) {
	keyTopicsJSON, marshalErr := json.Marshal(summary.KeyTopics)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal key topics: %w", marshalErr)
	}

	arg := sqlc.CreateConversationSummaryParams{
		UserID:          int32(summary.UserID),
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
		UserID:        int32(memory.UserID),
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
		UserID:     int32(memory.UserID),
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

func (r *Repository) SearchUserMemory(ctx context.Context, userID int, query string) ([]domain.UserMemory, error) {
	arg := sqlc.SearchUserMemoryParams{
		UserID:   int32(userID),
		CONCAT:   query,
		CONCAT_2: query,
		CONCAT_3: query,
	}

	results, err := db.NewReadExecutor[domain.UserMemory, sqlc.UserMemory](db.MainDatabase).FetchOrdered(ctx, r.queries.SearchUserMemory, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to search user memory", logger.Error(err))

		return nil, fmt.Errorf("failed to search user memory: %w", err)
	}

	return results, nil
}

func (r *Repository) GetSessionStats(ctx context.Context, userID int) (*domain.UserStats, error) {
	result, err := db.NewReadExecutor[domain.UserStats, sqlc.GetSessionStatsRow](db.MainDatabase).Fetch(ctx, r.queries.GetSessionStats, int32(userID))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session stats", logger.Error(err))

		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	return &result, nil
}

func (r *Repository) Close() error {
	return nil
}
