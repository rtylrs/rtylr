package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/caret/memory/internal/domain"
)

type repository interface {
	CreateChatSession(context.Context, *domain.ChatSession) (*domain.ChatSession, error)
	GetChatSession(context.Context, string, int) (*domain.ChatSession, error)
	UpdateChatSession(context.Context, string, int, string, string, string) (*domain.ChatSession, error)
	DeleteChatSession(context.Context, string, int) error
	CreateMessage(context.Context, *domain.ConversationMessage) (*domain.ConversationMessage, error)
	ListUserSessions(context.Context, int, int, int) ([]domain.ChatSession, error)
	GetSessionMessages(context.Context, string) ([]domain.ConversationMessage, error)
	GetContextMessages(context.Context, string) ([]domain.ConversationMessage, error)
	GetContextForSession(context.Context, string, int) ([]domain.ContextItem, error)
	CreateConversationSummary(context.Context, *domain.ConversationSummary) (*domain.ConversationSummary, error)
	StoreUserMemory(context.Context, *domain.UserMemory) (*domain.UserMemory, error)
	SearchUserMemory(context.Context, int, string) ([]domain.UserMemory, error)
	GetSessionStats(context.Context, int) (*domain.UserStats, error)
	ListProjects(context.Context) ([]domain.Project, error)
	GetProject(context.Context, string) (*domain.Project, error)
	SaveProject(context.Context, domain.Project, bool) (*domain.Project, error)
	ArchiveProject(context.Context, string) error
	SetChatProject(context.Context, string, int, string) error
	AssistantProfile(context.Context) (string, error)
	AssistantMemories(context.Context, string) ([]domain.UserMemory, error)
	ListMemories(context.Context) ([]domain.UserMemory, error)
	SaveMemory(context.Context, int, string) (*domain.UserMemory, error)
	DeleteMemory(context.Context, int) error
	ProjectConversationContext(context.Context, string, string) (string, error)
	SearchChats(context.Context, int, string, int) ([]domain.ChatSearchHit, error)
	SaveReplyFeedback(context.Context, int, domain.ReplyFeedback) error
	PendingReplyFeedback(context.Context, int, int) ([]domain.ReplyFeedback, error)
	ReplaceLessons(context.Context, int, []string, []int64) error
	UserLessons(context.Context, int) ([]string, error)
	SetChatMetadataNumber(context.Context, string, int, string, int) error
	SaveToolRoute(context.Context, int, string, string) error
	ToolRoutes(context.Context, int) ([]string, error)
}
