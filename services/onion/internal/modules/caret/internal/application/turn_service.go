package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/snips"
)

// MemoryService is the application-facing conversation port. HTTP handlers
// provide the concrete adapter, while turn preparation stays independent of
// transport and storage implementations.
type MemoryService interface {
	CreateSession(ctx context.Context, userID int, title string) (string, error)
	StoreMessage(ctx context.Context, sessionID string, userID int, role, content string, tokens int) (any, error)
	SummarizeSession(ctx context.Context, sessionID string, userID int) error
}

type AssistantContextService interface {
	AssistantContext(context.Context, string, int, string) (string, error)
}

type ImageMessageService interface {
	StoreMessageWithImages(context.Context, string, int, string, string, int, []snips.ImageInput) (any, error)
}

type MemorySessionAuthorizer interface {
	EnsureSessionOwnership(ctx context.Context, sessionID string, userID int) error
}

// TurnRequest contains the business input needed to prepare one Caret turn.
// HTTP aliases are resolved by the transport adapter before this type is
// created.
type TurnRequest struct {
	Prompt        string
	SessionID     string
	UserID        int64
	SystemContext string
	Mode          string
	PageContext   map[string]any
	Images        []snips.ImageInput
}

// PreparedTurn is the deterministic application result consumed by the
// streaming adapter. It keeps session state, prompt construction, and policy
// decisions together so each request has one source of truth.
type PreparedTurn struct {
	Request        TurnRequest
	Mode           CaretMode
	EnhancedPrompt string
	ResponsePolicy CaretResponsePolicy
	PreparedAt     time.Time
}

// NormalizeTurnRequest applies the same bounded, server-owned input rules
// before a request is persisted or included in a model prompt.
func NormalizeTurnRequest(req TurnRequest) TurnRequest {
	req.Prompt = SanitizeUserInput(req.Prompt, MaxUserPromptLen)
	req.SystemContext = BuildRequestSystemContext(req.SystemContext, req.PageContext)
	return req
}

// PrepareTurn owns the non-HTTP work required before an SSE stream starts:
// session ownership, user-message persistence, prompt construction, and one
// request policy decision.
func PrepareTurn(ctx context.Context, memory MemoryService, request TurnRequest, now time.Time) (PreparedTurn, error) {
	if memory == nil {
		return PreparedTurn{}, fmt.Errorf("memory service is not available")
	}

	request = NormalizeTurnRequest(request)
	storedPrompt := request.Prompt
	if normalizedPrompt, command, ok := NormalizeCaretSlashCommand(request.Prompt); ok {
		request.Prompt = normalizedPrompt
		request.Mode = string(command.Mode)
	}
	if request.UserID <= 0 {
		return PreparedTurn{}, fmt.Errorf("user id is required")
	}

	if strings.TrimSpace(request.SessionID) == "" {
		sessionID, err := memory.CreateSession(ctx, int(request.UserID), SessionTitleFromPrompt(storedPrompt))
		if err != nil {
			return PreparedTurn{}, fmt.Errorf("failed to create session: %w", err)
		}
		request.SessionID = sessionID
	} else if authorizer, ok := memory.(MemorySessionAuthorizer); ok {
		if err := authorizer.EnsureSessionOwnership(ctx, request.SessionID, int(request.UserID)); err != nil {
			return PreparedTurn{}, fmt.Errorf("session is not available to this user: %w", err)
		}
	}

	var storeErr error
	if len(request.Images) > 0 {
		if imageMemory, ok := memory.(ImageMessageService); ok {
			_, storeErr = imageMemory.StoreMessageWithImages(ctx, request.SessionID, int(request.UserID), "user", storedPrompt, len(storedPrompt)/4, request.Images)
		} else {
			storeErr = fmt.Errorf("image attachments are not supported by the memory service")
		}
	} else {
		_, storeErr = memory.StoreMessage(ctx, request.SessionID, int(request.UserID), "user", storedPrompt, len(storedPrompt)/4)
	}
	if storeErr != nil {
		return PreparedTurn{}, fmt.Errorf("failed to store message: %w", storeErr)
	}

	mode := ParseMode(request.Mode)
	enhancedPrompt := BuildCaretPrompt(PromptInput{
		Mode:           mode,
		CurrentRequest: request.Prompt,
		SystemContext:  request.SystemContext,
		Now:            now,
	})

	return PreparedTurn{
		Request:        request,
		Mode:           mode,
		EnhancedPrompt: enhancedPrompt,
		ResponsePolicy: BuildCaretResponsePolicy(mode, request.Prompt, request.SystemContext, request.PageContext),
		PreparedAt:     now,
	}, nil
}

func SessionTitleFromPrompt(prompt string) string {
	title := strings.Join(strings.Fields(prompt), " ")
	if title == "" {
		return "New chat"
	}

	const maxSessionTitleRunes = 80
	runes := []rune(title)
	if len(runes) <= maxSessionTitleRunes {
		return title
	}
	return string(runes[:maxSessionTitleRunes]) + "..."
}
