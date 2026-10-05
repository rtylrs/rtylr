package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/aicache"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"github.com/voxire/onion/internal/platform/snips"
)

func (h *GenerateTextStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Ctx(ctx)

	log.Info("Processing AI text stream request",
		logger.String("user_agent", r.UserAgent()),
		logger.String("remote_addr", r.RemoteAddr))

	req, err := h.parseAndValidateRequest(w, r)
	if err != nil {
		log.Error("Failed to parse request", logger.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)

		return
	}

	if err := h.validateRequest(ctx, req); err != nil {
		log.Error("Request validation failed", logger.Error(err))
		serverHttp.WriteOneKitError(ctx, w, err)

		return
	}

	if h.AiService == nil || !h.AiService.IsAvailable() {
		log.Warn("AI service not available")
		http.Error(w, "AI service not available", http.StatusServiceUnavailable)

		return
	}

	if h.MemoryService == nil {
		log.Error("Memory service not available")
		http.Error(w, "Memory service not available", http.StatusInternalServerError)

		return
	}
	// Quota gate. Use the authenticated identity (never the client-supplied
	// userId) so usage is metered against the real tenant.
	device := authPkg.DeviceCtx(ctx)
	var businessID, authUserID int64
	if device.User != nil {
		businessID = device.User.BusinessID
		authUserID = device.User.UserID
	}
	if authUserID <= 0 {
		if authUser, ok := authPkg.GetAuthUser(ctx); ok && authUser != nil {
			businessID = authUser.BusinessID
			authUserID = authUser.UserID
		}
	}
	if businessID <= 0 || authUserID <= 0 {
		log.Warn("AI request rejected without an authenticated business identity")
		http.Error(w, "Authentication is required", http.StatusUnauthorized)
		return
	}
	if projects, ok := h.MemoryService.(interface {
		ValidateProjectSelection(context.Context, string, string, int) error
	}); ok {
		if err := projects.ValidateProjectSelection(ctx, req.SessionID, req.ProjectID, int(authUserID)); err != nil {
			http.Error(w, "Selected project is not available for this chat", http.StatusBadRequest)
			return
		}
	}
	// Trust the authenticated identity over the client-supplied user_id for the
	// session/memory owner too — prevents one user reading another's history.
	if authUserID > 0 {
		req.UserID = authUserID
	}
	var reservation *quotaReservation
	if h.QuotaService != nil {
		allowed, plan, allowance, used, remaining, reserved := h.QuotaService.ReserveQuota(ctx, businessID, authUserID)
		if !allowed {
			log.Info("AI request blocked by quota",
				logger.Any("owner_business_id", businessID), logger.String("plan", plan))
			h.writeQuotaExceeded(ctx, w, plan, allowance, used, remaining)
			return
		}
		reservation = &quotaReservation{
			handler: h, ctx: ctx, businessID: businessID, userID: authUserID,
			reservedTokens: reserved,
		}
		defer reservation.settle(0)
	}

	requestNow := time.Now()
	prepared, setupErr := application.PrepareTurn(ctx, h.MemoryService, application.TurnRequest{
		Prompt:        req.Prompt,
		SessionID:     req.SessionID,
		UserID:        req.UserID,
		SystemContext: req.SystemContext,
		Mode:          "agent",
		PageContext:   req.PageContext,
		Images:        req.Images,
	}, requestNow)
	if setupErr != nil {
		log.Error("Failed to setup session and store message", logger.Error(setupErr))
		http.Error(w, "Failed to setup session", http.StatusInternalServerError)

		return
	}
	req.Mode = "agent"
	req.Prompt = prepared.Request.Prompt
	req.SessionID = prepared.Request.SessionID
	req.UserID = prepared.Request.UserID
	req.SystemContext = prepared.Request.SystemContext
	req.PageContext = prepared.Request.PageContext
	ctx = snips.WithConversation(ctx, req.SessionID)

	// Do not spend a model call on connectivity checks. A content-free prompt
	// gives a generative model room to invent a plausible business report even
	// though no business evidence was requested.
	if strings.EqualFold(strings.TrimSpace(req.Prompt), "ping") {
		h.setupSSEHeaders(w)
		h.streamDirectResponse(ctx, w, req, caretSmokeCheckReply)
		return
	}

	mode := application.ModeAgent
	savedContext := ""
	if memory, ok := h.MemoryService.(application.AssistantContextService); ok {
		savedContext, err = memory.AssistantContext(ctx, req.SessionID, int(authUserID), req.ProjectID)
		if err != nil {
			log.Error("Failed to load chat context", logger.Error(err))
			http.Error(w, "Failed to load chat context", http.StatusInternalServerError)
			return
		}
	}
	enhancedPrompt := application.BuildAssistantPrompt(req.Prompt, req.SystemContext, savedContext, req.PageContext, requestNow)
	ctx = snips.WithImages(ctx, req.Images)

	cachePlan := application.BuildCacheWritePlan(mode, application.TurnRequest{
		Prompt:        req.Prompt,
		SystemContext: req.SystemContext,
		PageContext:   req.PageContext,
	}, enhancedPrompt, businessID, authUserID, h.AiService.GetPrimaryProvider(), requestNow)

	cachePlan = application.DisableExactCachePlan(cachePlan, "conversation and project context change between turns")
	cachePlan.ProviderPlannerOnly = false
	cachePlan.ToolBudget = 6
	cachePlan.ContinuationBudget = 3
	req.EnhancedPrompt = enhancedPrompt

	h.setupSSEHeaders(w)
	clientGone := h.monitorClientDisconnect(r)

	if cachePlan.Enabled {
		if cached, ok := aicache.GetExact(ctx, cachePlan.Key); ok {
			log.Info("AI exact cache hit",
				logger.String("mode", cachePlan.Mode),
				logger.String("provider", cached.Provider),
				logger.String("cache_key", aicache.KeyPreview(cachePlan.Key)))
			h.streamCachedResponse(ctx, w, clientGone, req, cached, cachePlan, businessID, authUserID)
			return
		}
		aicache.Record(aicache.Event{
			Outcome:         aicache.OutcomeMiss,
			Mode:            cachePlan.Mode,
			Provider:        cachePlan.Provider,
			CacheKeyPreview: aicache.KeyPreview(cachePlan.Key),
			BusinessID:      businessID,
			UserID:          authUserID,
			PromptChars:     len(enhancedPrompt),
		})
	} else {
		aicache.Record(aicache.Event{
			Outcome:         aicache.OutcomeSkip,
			Mode:            cachePlan.Mode,
			Provider:        cachePlan.Provider,
			Reason:          cachePlan.Policy.Reason,
			BusinessID:      businessID,
			UserID:          authUserID,
			PromptChars:     len(enhancedPrompt),
			CacheKeyPreview: "",
		})
	}

	log.Info("Starting AI text generation",
		logger.String("enhanced_prompt_length", strconv.Itoa(len(enhancedPrompt))),
		logger.String("cache_mode", cachePlan.Mode),
		logger.String("cache_policy", cachePlan.Policy.Reason))

	tokenChan, err := h.AiService.GenerateTextStream(ctx, enhancedPrompt)
	if err != nil {
		log.Error("Failed to generate text stream", logger.Error(err))
		aicache.Record(aicache.Event{
			Outcome:     aicache.OutcomeError,
			Mode:        cachePlan.Mode,
			Provider:    cachePlan.Provider,
			Reason:      "provider stream failed to start",
			BusinessID:  businessID,
			UserID:      authUserID,
			PromptChars: len(enhancedPrompt),
		})
		h.writeErrorResponse(w, "Failed to generate response")

		return
	}

	h.streamTokensToClient(ctx, w, tokenChan, clientGone, req, businessID, authUserID, cachePlan, reservation)
}

func isCaretSmokeCheck(prompt string) bool {
	return application.ClassifyCaretRequest(prompt, "").IsSmokeCheck
}

func (h *GenerateTextStreamHandler) streamDirectResponse(ctx context.Context, w http.ResponseWriter, req *GenerateTextStreamRequest, response string) {
	messageID := newAssistantMessageID()
	if err := h.writeStreamPrelude(w, req.SessionID, messageID); err != nil {
		return
	}

	var fullResponse strings.Builder
	renderer := newStreamEventRenderer(w, flushResponse(w), messageID, &fullResponse, withStreamMode(application.ParseMode(req.Mode)))
	if err := renderer.AppendText(response); err != nil {
		return
	}
	if err := renderer.FlushPending(); err != nil {
		return
	}
	if err := h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String()); err != nil {
		h.writeErrorResponse(w, "Failed to persist response")
		return
	}
	if err := renderer.Finish(); err != nil {
		return
	}
}
