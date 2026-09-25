package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/memory/internal/domain"
	"github.com/voxire/onion/internal/modules/memory/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const memoryTTL = 30 * time.Second

const (
	maxSummaryInputRunes = 12000
)

var ErrChatSessionNotFound = infrastructure.ErrChatSessionNotFound

func invalidateMemoryScope(ctx context.Context, userID int) {
	cacher.InvalidateScope(ctx, cacheKeys.UserPrefix(int64(userID)))
}

type Service struct {
	repository *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repository: repo,
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
	if s == nil || s.repository == nil {
		return nil, errors.New("memory repository not available")
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("session id is required")
	}
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "user" && role != "assistant" && role != "system" {
		return nil, errors.New("invalid message role")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("message content is required")
	}
	if tokensUsed < 0 {
		tokensUsed = 0
	}
	if _, err := s.repository.GetChatSession(ctx, sessionID, userID); err != nil {
		return nil, err
	}
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

// EnsureSessionOwnership verifies the session before any caller can append a
// message or request context for it. The query is user-scoped, so an unknown or
// cross-user session has the same result and does not disclose ownership.
func (s *Service) EnsureSessionOwnership(ctx context.Context, sessionID string, userID int) error {
	if s == nil || s.repository == nil {
		return errors.New("memory repository not available")
	}
	if strings.TrimSpace(sessionID) == "" || userID <= 0 {
		return errors.New("valid session and user are required")
	}
	_, err := s.repository.GetChatSession(ctx, sessionID, userID)
	return err
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
	if s == nil || s.repository == nil {
		return errors.New("memory repository not available")
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

	// Conversation content is deliberately never sent to a provider from the
	// memory lifecycle. Keep the summary API useful with a deterministic,
	// content-free record; provider requests can still use the current turn and
	// explicit server-side tools without receiving historical messages.
	summaryData := deterministicSummaryData(messages)

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

	// Topics help the next response understand this conversation, but they are
	// not durable user preferences. Durable memory is written only through the
	// explicit, approval-backed memory.remember tool; otherwise a model's topic
	// guess could silently become a future business fact.

	return nil
}

func deterministicSummaryData(messages []domain.ConversationMessage) *SummaryData {
	userMessages := 0
	assistantMessages := 0
	for _, message := range messages {
		switch strings.ToLower(strings.TrimSpace(message.Role)) {
		case "user":
			userMessages++
		case "assistant":
			assistantMessages++
		}
	}

	return &SummaryData{
		Summary:         fmt.Sprintf("This session contains %d user request(s) and %d assistant response(s). Message content remains server-side.", userMessages, assistantMessages),
		KeyTopics:       []string{"conversation"},
		Sentiment:       "neutral",
		ImportanceScore: 0.5,
	}
}

func truncateSummaryRunes(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes]) + "..."
}

func (s *Service) identifyMemories(_ []domain.ConversationMessage, summaryData *SummaryData) []MemoryItem {
	memories := make([]MemoryItem, 0, len(summaryData.KeyTopics))

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
		line = strings.Trim(line, "`*_ ")
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "summary":
			summaryData.Summary = truncateSummaryRunes(value, maxSummaryInputRunes/2)
		case "key topics", "key_topics":
			topics := value
			seen := make(map[string]struct{})
			for _, topic := range strings.Split(topics, ",") {
				topic = strings.TrimSpace(topic)
				if topic == "" || len(summaryData.KeyTopics) >= 5 {
					continue
				}
				key := strings.ToLower(topic)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				summaryData.KeyTopics = append(summaryData.KeyTopics, truncateSummaryRunes(topic, 80))
			}
		case "sentiment":
			sentiment := strings.ToLower(value)
			switch sentiment {
			case "positive", "negative", "neutral":
				summaryData.Sentiment = sentiment
			}
		case "importance", "importance score", "importance_score":
			if score, err := strconv.ParseFloat(value, 64); err == nil {
				if score < 0 {
					score = 0
				}
				if score > 1 {
					score = 1
				}
				summaryData.ImportanceScore = score
			}
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

func setSessionMetadataFlag(raw, key string, value bool) (string, error) {
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
