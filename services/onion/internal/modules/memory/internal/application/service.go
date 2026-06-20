package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	"github.com/voxire/onion/internal/modules/memory/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/snips"
)

const memoryTTL = 30 * time.Second

var ErrChatSessionNotFound = infrastructure.ErrChatSessionNotFound

func invalidateMemoryScope(ctx context.Context, userID int) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.UserPrefix(int64(userID)))
	}
}

var ErrAIClientNotAvailable = errors.New("AI client not available for summarization")

type Service struct {
	repository *infrastructure.Repository
	aiClient   *snips.Client
}

func NewService(ctx context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	aiClient, err := snips.NewClient(ctx)
	if err != nil {
		logger.Ctx(ctx).Debug("AI client not available for memory summarization", logger.Error(err))
		aiClient = nil
	}

	return &Service{
		repository: repo,
		aiClient:   aiClient,
	}
}

func (s *Service) CreateSession(ctx context.Context, userID int, title string) (*domain.ChatSession, error) {
	sessionID := uuid.New().String()

	session := &domain.ChatSession{
		UserID:    userID,
		SessionID: sessionID,
		Title:     title,
		Metadata:  "{}",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	defer invalidateMemoryScope(ctx, userID)
	return s.repository.CreateChatSession(ctx, session)
}

// ListUserSessions returns the user's chat sessions ordered by most recent activity,
// applying paginated defaults (limit 20, max 100) per the project pagination rules.
func (s *Service) ListUserSessions(ctx context.Context, userID, limit, offset int) ([]domain.ChatSession, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	key := cacheKeys.Scoped("user", int64(userID), fmt.Sprintf("memory_sessions|l=%d|o=%d", limit, offset))
	return cacher.Of[[]domain.ChatSession]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.ChatSession, error) {
			return s.repository.ListUserSessions(ctx, userID, limit, offset)
		},
		cacher.WithExpiry(ctx, memoryTTL),
	)
}

func (s *Service) StoreMessage(ctx context.Context, sessionID string, userID int, role, content string, tokensUsed int) (*domain.ConversationMessage, error) {
	defer invalidateMemoryScope(ctx, userID)
	message := &domain.ConversationMessage{
		SessionID:  sessionID,
		UserID:     userID,
		Role:       role,
		Content:    content,
		TokensUsed: tokensUsed,
		CreatedAt:  time.Now(),
	}

	return s.repository.CreateMessage(ctx, message)
}

func (s *Service) GetConversationContext(ctx context.Context, sessionID string, userID int) (*domain.ConversationContext, error) {
	log := logger.Ctx(ctx)

	log.Info("Getting conversation context",
		logger.String("session_id", sessionID),
		logger.Int("user_id", userID))

	contextData, err := s.repository.GetContextForSession(ctx, sessionID, userID)
	if err != nil {
		log.Error("Failed to get conversation context from repository", logger.Error(err),
			logger.String("session_id", sessionID),
			logger.Int("user_id", userID))

		return nil, fmt.Errorf("failed to get conversation context: %w", err)
	}

	log.Info("Context data retrieved from repository",
		logger.String("session_id", sessionID),
		logger.Int("data_items", len(contextData)))

	conversationContext := &domain.ConversationContext{
		SessionID: sessionID,
		UserID:    userID,
		Messages:  []domain.ConversationMessage{},
		Summary:   "",
		Memories:  []domain.UserMemory{},
	}

	for i := range contextData {
		item := contextData[i]
		switch item.Type {
		case "message":
			message := domain.ConversationMessage{
				Role:    item.Role,
				Content: item.Content,
			}
			if item.CreatedAt != nil {
				message.CreatedAt = *item.CreatedAt
			}

			conversationContext.Messages = append(conversationContext.Messages, message)
		case "summary":
			conversationContext.Summary = item.Summary
		case "memory":
			conversationContext.Memories = append(conversationContext.Memories, domain.UserMemory{
				MemoryType: item.MemoryType,
				Key:        item.MemoryKey,
				Value:      item.MemoryValue,
			})
		}
	}

	log.Info("Conversation context built successfully",
		logger.String("session_id", sessionID),
		logger.Int("messages", len(conversationContext.Messages)),
		logger.Int("memories", len(conversationContext.Memories)))

	return conversationContext, nil
}

func (s *Service) SummarizeSession(ctx context.Context, sessionID string, userID int) error {
	if s.aiClient == nil {
		return ErrAIClientNotAvailable
	}

	// Verify the session belongs to this user before reading its messages —
	// GetSessionMessages is keyed by session id alone, so without this guard a
	// user could summarize (and surface) another user's conversation.
	if _, err := s.repository.GetChatSession(ctx, sessionID, userID); err != nil {
		return err
	}

	messages, err := s.repository.GetSessionMessages(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session messages: %w", err)
	}

	if len(messages) < 3 {
		return nil
	}

	var conversationText strings.Builder
	for _, msg := range messages {
		_, _ = fmt.Fprintf(&conversationText, "%s: %s\n", msg.Role, msg.Content)
	}

	prompt := fmt.Sprintf(`Please provide a concise summary of this conversation in 2-3 sentences. Also identify 3-5 key topics discussed and assess the overall sentiment (positive, negative, or neutral).

Conversation:
%s

Summary:`, conversationText.String())

	summary, err := s.aiClient.CompleteText(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	summaryData := s.parseSummaryResponse(summary)

	conversationSummary := &domain.ConversationSummary{
		UserID:          userID,
		SessionID:       sessionID,
		Summary:         summaryData.Summary,
		KeyTopics:       summaryData.KeyTopics,
		Sentiment:       summaryData.Sentiment,
		ImportanceScore: summaryData.ImportanceScore,
		CreatedAt:       time.Now(),
	}

	_, err = s.repository.CreateConversationSummary(ctx, conversationSummary)
	if err != nil {
		return fmt.Errorf("failed to store conversation summary: %w", err)
	}

	s.extractAndStoreMemories(ctx, userID, sessionID, messages, summaryData)

	return nil
}

func (s *Service) extractAndStoreMemories(ctx context.Context, userID int, sessionID string, messages []domain.ConversationMessage, summaryData *SummaryData) {
	memories := s.identifyMemories(messages, summaryData)

	for _, memory := range memories {
		userMemory := &domain.UserMemory{
			UserID:        userID,
			MemoryType:    memory.Type,
			Key:           memory.Key,
			Value:         memory.Value,
			Confidence:    memory.Confidence,
			SourceSession: sessionID,
			CreatedAt:     time.Now(),
		}

		_, err := s.repository.StoreUserMemory(ctx, userMemory)
		if err != nil {
			logger.Ctx(ctx).Warn("Failed to store user memory",
				logger.Error(err),
				logger.String("memory_key", memory.Key))
		}
	}
}

func (s *Service) identifyMemories(_ []domain.ConversationMessage, summaryData *SummaryData) []MemoryItem {
	memories := make([]MemoryItem, 0, len(summaryData.KeyTopics))

	if summaryData.Sentiment != "" {
		memories = append(memories, MemoryItem{
			Type:       "preference",
			Key:        "conversation_style",
			Value:      fmt.Sprintf("User prefers %s conversations", summaryData.Sentiment),
			Confidence: 0.8,
		})
	}

	for _, topic := range summaryData.KeyTopics {
		memories = append(memories, MemoryItem{
			Type:       "interest",
			Key:        "topic_" + strings.ToLower(strings.ReplaceAll(topic, " ", "_")),
			Value:      "User is interested in " + topic,
			Confidence: 0.7,
		})
	}

	return memories
}

func (s *Service) parseSummaryResponse(response string) *SummaryData {
	lines := strings.Split(response, "\n")

	summaryData := &SummaryData{
		Summary:         "",
		KeyTopics:       []string{},
		Sentiment:       "neutral",
		ImportanceScore: 1.0,
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Summary:") {
			summaryData.Summary = strings.TrimPrefix(line, "Summary:")
		} else if strings.HasPrefix(line, "Key topics:") {
			topics := strings.TrimPrefix(line, "Key topics:")
			summaryData.KeyTopics = strings.Split(topics, ",")
		} else if strings.HasPrefix(line, "Sentiment:") {
			sentiment := strings.TrimPrefix(line, "Sentiment:")
			summaryData.Sentiment = strings.TrimSpace(sentiment)
		}
	}

	return summaryData
}

func (s *Service) SearchMemories(ctx context.Context, userID int, query string) ([]domain.UserMemory, error) {
	return s.repository.SearchUserMemory(ctx, userID, query)
}

func (s *Service) StoreUserMemory(ctx context.Context, userID int, memoryType, key, value, sourceSession string, confidence float64) (*domain.UserMemory, error) {
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	memoryType = strings.TrimSpace(memoryType)
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if memoryType == "" {
		return nil, errors.New("memory type is required")
	}
	if key == "" {
		return nil, errors.New("memory key is required")
	}
	if value == "" {
		return nil, errors.New("memory value is required")
	}
	if confidence <= 0 {
		confidence = 1
	}
	if confidence > 1 {
		confidence = 1
	}

	defer invalidateMemoryScope(ctx, userID)
	return s.repository.StoreUserMemory(ctx, &domain.UserMemory{
		UserID:        userID,
		MemoryType:    memoryType,
		Key:           key,
		Value:         value,
		Confidence:    confidence,
		SourceSession: strings.TrimSpace(sourceSession),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
}

func (s *Service) RenameSession(ctx context.Context, sessionID string, userID int, title string) (*domain.ChatSession, error) {
	session, err := s.repository.GetChatSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	defer invalidateMemoryScope(ctx, userID)
	return s.repository.UpdateChatSession(ctx, sessionID, userID, title, session.Summary, session.Metadata)
}

func (s *Service) SetSessionPinned(ctx context.Context, sessionID string, userID int, pinned bool) (*domain.ChatSession, error) {
	session, err := s.repository.GetChatSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	metadata, err := setSessionMetadataFlag(session.Metadata, "pinned", pinned)
	if err != nil {
		return nil, err
	}

	defer invalidateMemoryScope(ctx, userID)
	return s.repository.UpdateChatSession(ctx, sessionID, userID, session.Title, session.Summary, metadata)
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string, userID int) error {
	defer invalidateMemoryScope(ctx, userID)
	return s.repository.DeleteChatSession(ctx, sessionID, userID)
}

func (s *Service) GetUserStats(ctx context.Context, userID int) (*domain.UserStats, error) {
	return cacher.Of[*domain.UserStats]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped("user", int64(userID), "memory_stats"),
		func(ctx context.Context) (*domain.UserStats, error) {
			stats, err := s.repository.GetSessionStats(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user stats: %w", err)
			}
			return &domain.UserStats{
				TotalMessages:       stats.TotalMessages,
				TotalSessions:       stats.TotalSessions,
				AvgTokensPerMessage: stats.AvgTokensPerMessage,
				TotalTokensUsed:     stats.TotalTokensUsed,
			}, nil
		},
		cacher.WithExpiry(ctx, memoryTTL),
	)
}

func (s *Service) Close() error {
	if s.repository != nil {
		return s.repository.Close()
	}

	return nil
}

func setSessionMetadataFlag(raw string, key string, value bool) (string, error) {
	metadata := map[string]interface{}{}
	if strings.TrimSpace(raw) != "" && strings.TrimSpace(raw) != "null" {
		if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
			return "", fmt.Errorf("failed to parse session metadata: %w", err)
		}
	}

	metadata[key] = value

	encoded, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to encode session metadata: %w", err)
	}

	return string(encoded), nil
}

type SummaryData struct {
	Summary         string   `json:"summary"`
	KeyTopics       []string `json:"key_topics"`
	Sentiment       string   `json:"sentiment"`
	ImportanceScore float64  `json:"importance_score"`
}

type MemoryItem struct {
	Type       string  `json:"type"`
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}
