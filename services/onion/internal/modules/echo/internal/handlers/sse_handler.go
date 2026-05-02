package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/snips"
	"github.com/voxire/onion/internal/modules/echo/internal/application"
	"github.com/voxire/onion/internal/modules/memory"
)

const caretSystemPreamble = "You are Caret, an AI business assistant for rtylr."

var ErrInvalidRequestBody = errors.New("invalid request body")

func GenerateTextStreamRoutes() []string {
	return []string{
		"/ai/generate-text-stream",
	}
}

type GenerateTextStreamHandler struct {
	AiService     *application.Service
	MemoryService *memory.Module
}

type GenerateTextStreamRequest struct {
	Prompt        string `json:"prompt"`
	SessionID     string `json:"sessionId"`
	UserID        int64  `json:"userId"`
	SystemContext string `json:"system_context"` // optional — app context injected as system message
}

type GenerateTextStreamResponse struct {
	Response   string `json:"response"`
	IsComplete bool   `json:"isComplete"`
	SessionID  string `json:"sessionId"`
}

func (h *GenerateTextStreamHandler) Validate(_ context.Context, req *GenerateTextStreamRequest) *errorModels.ServerError {
	if req.Prompt == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Prompt is required",
		}
	}

	if req.UserID <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "User ID is required",
		}
	}

	return nil
}

func (h *GenerateTextStreamHandler) Serve(_ context.Context, _ *GenerateTextStreamRequest) (GenerateTextStreamResponse, *errorModels.ServerError) {
	if !h.AiService.IsAvailable() {
		return GenerateTextStreamResponse{}, &errorModels.ServerError{
			Code:    http.StatusServiceUnavailable,
			Message: "AI service not available",
		}
	}

	return GenerateTextStreamResponse{}, nil
}

func (h *GenerateTextStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.Ctx(ctx)

	log.Info("Processing AI text stream request",
		logger.String("user_agent", r.UserAgent()),
		logger.String("remote_addr", r.RemoteAddr))

	req, err := h.parseAndValidateRequest(r)
	if err != nil {
		log.Error("Failed to parse request", logger.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)

		return
	}

	if err := h.validateRequest(ctx, req); err != nil {
		log.Error("Request validation failed", logger.Error(err))
		http.Error(w, err.Message, err.Code)

		return
	}

	if !h.AiService.IsAvailable() {
		log.Warn("AI service not available")
		http.Error(w, "AI service not available", http.StatusServiceUnavailable)

		return
	}

	if h.MemoryService == nil || h.MemoryService.GetService() == nil {
		log.Error("Memory service not available")
		http.Error(w, "Memory service not available", http.StatusInternalServerError)

		return
	}

	if setupErr := h.setupSessionAndStoreMessage(ctx, req); setupErr != nil {
		log.Error("Failed to setup session and store message", logger.Error(setupErr))
		http.Error(w, "Failed to setup session", http.StatusInternalServerError)

		return
	}

	enhancedPrompt, contextErr := h.buildEnhancedPromptWithContext(ctx, req)
	if contextErr != nil {
		log.Error("Failed to get conversation context", logger.Error(contextErr))
		http.Error(w, "Failed to get context", http.StatusInternalServerError)

		return
	}

	log.Info("Starting AI text generation",
		logger.String("enhanced_prompt_length", strconv.Itoa(len(enhancedPrompt))))

	h.setupSSEHeaders(w)
	clientGone := h.monitorClientDisconnect(r)

	tokenChan, err := h.AiService.GenerateTextStream(ctx, enhancedPrompt)
	if err != nil {
		log.Error("Failed to generate text stream", logger.Error(err))
		h.writeErrorResponse(w, "Failed to generate response")

		return
	}

	h.streamTokensToClient(ctx, w, tokenChan, clientGone, req)
}

func (h *GenerateTextStreamHandler) parseAndValidateRequest(r *http.Request) (*GenerateTextStreamRequest, error) {
	var req GenerateTextStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, ErrInvalidRequestBody
	}

	return &req, nil
}

func (h *GenerateTextStreamHandler) validateRequest(ctx context.Context, req *GenerateTextStreamRequest) *errorModels.ServerError {
	return h.Validate(ctx, req)
}

func (h *GenerateTextStreamHandler) setupSessionAndStoreMessage(ctx context.Context, req *GenerateTextStreamRequest) error {
	log := logger.Ctx(ctx)

	log.Info("Setting up session and storing message",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)),
		logger.String("prompt_length", strconv.Itoa(len(req.Prompt))))

	if req.SessionID == "" {
		log.Info("Creating new session for user", logger.Int("user_id", int(req.UserID)))

		session, err := h.MemoryService.GetService().CreateSession(ctx, int(req.UserID), "Auto-generated session")
		if err != nil {
			log.Error("Failed to create session", logger.Error(err), logger.Int("user_id", int(req.UserID)))

			return fmt.Errorf("failed to create session: %w", err)
		}

		req.SessionID = session.SessionID
		log.Info("Session created successfully", logger.String("session_id", req.SessionID))
	} else {
		log.Info("Using existing session", logger.String("session_id", req.SessionID))
	}

	log.Info("Storing user message",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)),
		logger.String("role", "user"))

	_, err := h.MemoryService.GetService().StoreMessage(ctx, req.SessionID, int(req.UserID), "user", req.Prompt, len(req.Prompt)/4)
	if err != nil {
		log.Error("Failed to store message", logger.Error(err),
			logger.String("session_id", req.SessionID),
			logger.Int("user_id", int(req.UserID)))

		return fmt.Errorf("failed to store message: %w", err)
	}

	log.Info("Message stored successfully")

	return nil
}

func (h *GenerateTextStreamHandler) buildEnhancedPromptWithContext(ctx context.Context, req *GenerateTextStreamRequest) (string, error) {
	logger.Ctx(ctx).Info("Building enhanced prompt with context",
		logger.String("session_id", req.SessionID),
		logger.Int("user_id", int(req.UserID)))

	enhancedPrompt := h.MemoryService.BuildEnhancedPrompt(ctx, req.SessionID, int(req.UserID), req.Prompt)

	if req.SystemContext == "" {
		return enhancedPrompt, nil
	}

	var sb strings.Builder
	sb.WriteString(caretSystemPreamble)
	sb.WriteString(" Today is ")
	sb.WriteString(time.Now().Format("2006-01-02"))
	sb.WriteString(".\nCurrent application context:\n")
	sb.WriteString(req.SystemContext)
	sb.WriteString("\n\n---\n\n")
	sb.WriteString(enhancedPrompt)

	return sb.String(), nil
}

func (h *GenerateTextStreamHandler) setupSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")
}

func (h *GenerateTextStreamHandler) monitorClientDisconnect(r *http.Request) chan bool {
	clientGone := make(chan bool)

	go func() {
		<-r.Context().Done()

		clientGone <- true
	}()

	return clientGone
}

func (h *GenerateTextStreamHandler) writeErrorResponse(w http.ResponseWriter, errorMsg string) {
	if _, writeErr := fmt.Fprintf(w, "event: error\ndata: %s\n\n", jsonEscape(errorMsg)); writeErr != nil {
		logger.Global().Error("Failed to write error to SSE response", logger.Error(writeErr))
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (h *GenerateTextStreamHandler) streamTokensToClient(ctx context.Context, w http.ResponseWriter, tokenChan <-chan snips.StreamToken, clientGone chan bool, req *GenerateTextStreamRequest) {
	var fullResponse strings.Builder

	for {
		select {
		case token := <-tokenChan:
			if token.Error != "" {
				h.writeErrorResponse(w, token.Error)

				return
			}

			if err := h.processAndStreamToken(w, token, &fullResponse); err != nil {
				continue
			}

			if token.Done {
				if fullResponse.Len() > 0 {
					go h.storeAIResponse(ctx, req.SessionID, int(req.UserID), fullResponse.String())
				}

				return
			}

		case <-clientGone:
			return

		case <-ctx.Done():
			return
		}
	}
}

func (h *GenerateTextStreamHandler) processAndStreamToken(w http.ResponseWriter, token snips.StreamToken, fullResponse *strings.Builder) error {
	if token.Error != "" {
		return nil
	}

	data, err := json.Marshal(token)
	if err != nil {
		h.writeErrorResponse(w, "Failed to marshal token")

		return err
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		logger.Global().Error("Failed to write SSE data to response", logger.Error(err))

		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	if token.Text != "" {
		fullResponse.WriteString(token.Text)
	}

	return nil
}

// buildEnhancedPromptWithContext now delegates to the memory module so the
// prompt-assembly rules live next to the data they read. The
// conversationContext argument is retained as a no-op so the existing call
// site signature does not change.
func (h *GenerateTextStreamHandler) buildEnhancedPrompt(userPrompt string, _ interface{}) string {
	return "Current request: " + userPrompt
}

func (h *GenerateTextStreamHandler) storeAIResponse(ctx context.Context, sessionID string, userID int, response string) {
	_, err := h.MemoryService.GetService().StoreMessage(ctx, sessionID, userID, "assistant", response, len(response)/4)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to store AI response",
			logger.Error(err),
			logger.String("session_id", sessionID),
			logger.Int("user_id", userID))
	}

	go func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if summarizeErr := h.MemoryService.GetService().SummarizeSession(timeoutCtx, sessionID, userID); summarizeErr != nil {
			logger.Ctx(ctx).Warn("Failed to summarize session",
				logger.Error(summarizeErr),
				logger.String("session_id", sessionID),
				logger.Int("user_id", userID))
		}
	}()
}

func jsonEscape(s string) string {
	return fmt.Sprintf("%q", s)
}
