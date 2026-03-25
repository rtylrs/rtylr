package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool    *pgxpool.Pool
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
		Title:     pgtype.Text{String: session.Title, Valid: session.Title != ""},
		Metadata:  []byte(session.Metadata),
	}

	result, err := db.NewWriteExecutor[domain.ChatSession, sqlc.ChatSession](db.MainDatabase).Insert(ctx, r.queries.CreateChatSession, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	return &result, nil
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
		SessionID: sessionID,
		UserID:    int32(userID),
		Title:     pgtype.Text{String: title, Valid: title != ""},
		Summary:   pgtype.Text{String: summary, Valid: summary != ""},
		Metadata:  []byte(metadata),
	}

	result, err := db.NewWriteExecutor[domain.ChatSession, sqlc.ChatSession](db.MainDatabase).Update(ctx, r.queries.UpdateChatSession, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update chat session", logger.Error(err))

		return nil, fmt.Errorf("failed to update chat session: %w", err)
	}

	return &result, nil
}

func (r *Repository) CreateMessage(ctx context.Context, message *domain.ConversationMessage) (*domain.ConversationMessage, error) {
	arg := sqlc.CreateMessageParams{
		SessionID:  message.SessionID,
		UserID:     int32(message.UserID),
		Role:       message.Role,
		Content:    message.Content,
		TokensUsed: pgtype.Int4{Int32: int32(message.TokensUsed), Valid: true},
		Metadata:   []byte("{}"),
	}

	result, err := db.NewWriteExecutor[domain.ConversationMessage, sqlc.ConversationMessage](db.MainDatabase).Insert(ctx, r.queries.CreateMessage, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create message", logger.Error(err))

		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return &result, nil
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
		SessionID: sessionID,
		UserID:    int32(userID),
	}

	results, err := db.NewReadExecutor[domain.ContextItem, sqlc.GetContextForSessionRow](db.MainDatabase).FetchOrdered(ctx, r.queries.GetContextForSession, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to query context", logger.Error(err))

		return nil, fmt.Errorf("failed to query context: %w", err)
	}

	return results, nil
}

func (r *Repository) CreateConversationSummary(ctx context.Context, summary *domain.ConversationSummary) (*domain.ConversationSummary, error) {
	arg := sqlc.CreateConversationSummaryParams{
		UserID:          int32(summary.UserID),
		SessionID:       summary.SessionID,
		Summary:         summary.Summary,
		KeyTopics:       summary.KeyTopics,
		Sentiment:       pgtype.Text{String: summary.Sentiment, Valid: summary.Sentiment != ""},
		ImportanceScore: pgtype.Float8{Float64: summary.ImportanceScore, Valid: true},
	}

	result, err := db.NewWriteExecutor[domain.ConversationSummary, sqlc.ConversationSummary](db.MainDatabase).Insert(ctx, r.queries.CreateConversationSummary, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create conversation summary", logger.Error(err))

		return nil, fmt.Errorf("failed to create conversation summary: %w", err)
	}

	return &result, nil
}

func (r *Repository) StoreUserMemory(ctx context.Context, memory *domain.UserMemory) (*domain.UserMemory, error) {
	arg := sqlc.StoreUserMemoryParams{
		UserID:        int32(memory.UserID),
		MemoryType:    memory.MemoryType,
		Key:           memory.Key,
		Value:         memory.Value,
		Confidence:    pgtype.Float8{Float64: memory.Confidence, Valid: true},
		SourceSession: pgtype.Text{String: memory.SourceSession, Valid: memory.SourceSession != ""},
	}

	result, err := db.NewWriteExecutor[domain.UserMemory, sqlc.UserMemory](db.MainDatabase).Insert(ctx, r.queries.StoreUserMemory, arg)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store user memory", logger.Error(err))

		return nil, fmt.Errorf("failed to store user memory: %w", err)
	}

	return &result, nil
}

func (r *Repository) SearchUserMemory(ctx context.Context, userID int, query string) ([]domain.UserMemory, error) {
	arg := sqlc.SearchUserMemoryParams{
		UserID: int32(userID),
		Key:    query,
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
