package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/snips"
)

// QuotaService reserves AI tokens against the tenant's monthly allowance and
// reconciles actual consumption afterwards. Implemented by the aiquota module;
// kept as a local interface (primitive signatures) so echo doesn't depend on
// its internals. ReserveQuota atomically debits an estimate before the provider
// call — parallel requests can no longer all pass an exhausted bucket.
type QuotaService interface {
	ReserveQuota(ctx context.Context, businessID, userID int64) (allowed bool, planSlug string, allowanceTokens, usedTokens, remainingTokens, reservedTokens int64)
	ReconcileUsage(ctx context.Context, businessID, userID, actualTokens, reservedTokens int64)
}

// quotaReservation settles exactly once, including early exits where no
// provider call occurs. A reservation must not become a permanent debit just
// because setup, cache lookup, or stream startup returned early.
type quotaReservation struct {
	once           sync.Once
	handler        *GenerateTextStreamHandler
	ctx            context.Context
	businessID     int64
	userID         int64
	reservedTokens int64
}

func (q *quotaReservation) settle(actualTokens int64) {
	if q == nil {
		return
	}
	q.once.Do(func() {
		q.handler.recordUsage(q.ctx, q.businessID, q.userID, actualTokens, q.reservedTokens)
	})
}

const (
	caretSmokeCheckReply = "Caret is working. What would you like help with?"
)

const (
	cachedReplayTinyResponseRunes = 12
	cachedReplayMinChunkRunes     = 8
	cachedReplayMinTailRunes      = 8
	cachedReplaySoftChunkRunes    = 12
	cachedReplayMaxChunkRunes     = 24
	cachedReplayMinDelay          = 95 * time.Millisecond
	cachedReplayMaxDelay          = 650 * time.Millisecond
	cachedReplayTargetDuration    = 8500 * time.Millisecond
	unexpectedStreamClosureError  = "Response stream ended before completion"
	sseHeartbeatInterval          = 15 * time.Second
)

var ErrInvalidRequestBody = errors.New("invalid request body")

func GenerateTextStreamRoutes() []string {
	return []string{
		"/ai/generate-text-stream",
	}
}

type GenerateTextStreamHandler struct {
	AiService      AIStreamService
	MemoryService  application.MemoryService
	QuotaService   QuotaService
	AnalystService AnalystService
	ToolExecutors  []ToolExecutor
	summaryMu      sync.Mutex
	summaryRunning map[string]struct{}
}

type AIStreamService interface {
	IsAvailable() bool
	GetPrimaryProvider() string
	GenerateTextStream(context.Context, string) (<-chan snips.StreamToken, error)
}

type GenerateTextStreamRequest struct {
	ProjectID string             `json:"project_id,omitempty"`
	Images    []snips.ImageInput `json:"images,omitempty"`
	// Files are text attachments (CSV, notes, JSON) sent as their text.
	Files          []application.AttachedFile `json:"files,omitempty"`
	EnhancedPrompt string                     `json:"-"`
	Prompt         string                     `json:"prompt"`
	SessionID      string                     `json:"session_id"`
	UserID         int64                      `json:"user_id"`
	SystemContext  string                     `json:"system_context"` // optional — app context injected as system message
	Mode           string                     `json:"mode"`           // optional — "analyst" routes through the NL-to-SQL layer
	PageContext    map[string]any             `json:"page_context,omitempty"`
	// Timezone is the person's IANA timezone, from their browser.
	Timezone string `json:"timezone,omitempty"`
}

// UnmarshalJSON keeps older Caret clients working while maintaining a single
// snake_case representation in the request type and in generated responses.
func (r *GenerateTextStreamRequest) UnmarshalJSON(data []byte) error {
	fields := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	decode := func(dst any, names ...string) error {
		for _, name := range names {
			raw, ok := fields[name]
			if !ok {
				continue
			}
			return json.Unmarshal(raw, dst)
		}
		return nil
	}

	var parsed GenerateTextStreamRequest
	if err := decode(&parsed.Prompt, "prompt"); err != nil {
		return err
	}
	if err := decode(&parsed.SessionID, "session_id", "sessionId"); err != nil {
		return err
	}
	if err := decode(&parsed.UserID, "user_id", "userId"); err != nil {
		return err
	}
	if err := decode(&parsed.SystemContext, "system_context"); err != nil {
		return err
	}
	if err := decode(&parsed.Mode, "mode"); err != nil {
		return err
	}
	if err := decode(&parsed.PageContext, "page_context", "pageContext"); err != nil {
		return err
	}
	if err := decode(&parsed.ProjectID, "project_id"); err != nil {
		return err
	}
	if err := decode(&parsed.Images, "images"); err != nil {
		return err
	}
	if err := decode(&parsed.Files, "files"); err != nil {
		return err
	}
	if err := decode(&parsed.Timezone, "timezone"); err != nil {
		return err
	}
	*r = parsed
	return nil
}

type GenerateTextStreamResponse struct {
	Response   string `json:"response"`
	IsComplete bool   `json:"is_complete"`
	SessionID  string `json:"session_id"`
}

func (h *GenerateTextStreamHandler) Validate(_ context.Context, req *GenerateTextStreamRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "Request is required")
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Prompt is required")
	}

	if req.UserID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "User ID is required")
	}

	if err := snips.ValidateImages(req.Images); err != nil {
		return onekiterrors.New(http.StatusBadRequest, "Invalid image attachment")
	}
	if err := application.ValidateAttachedFiles(req.Files); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}
