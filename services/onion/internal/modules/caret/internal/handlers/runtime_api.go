package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubcaretv1 "github.com/voxire/voxie/gen/go/hub/caret/v1"
)

// RuntimeAPI serves hub.caret.v1.CaretRuntimeService. It is the typed
// transport in front of the existing handlers, which keep the tool and
// approval logic.
type RuntimeAPI struct {
	Tools    *ListToolsHandler
	Status   *ProviderStatusHandler
	Decision *ToolDecisionHandler
	// Learner stores reply feedback; AI distills it into lessons.
	Learner application.LessonLearner
	AI      application.TextCompleter
}

// RateReply records the caller's verdict on a reply in one of their chats and,
// when it is worth it, starts a background pass that learns from it.
func (a *RuntimeAPI) RateReply(ctx context.Context, req *hubcaretv1.RateReplyRequest) (*hubcaretv1.RateReplyResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 {
		return nil, onekiterrors.PermissionDenied()
	}
	if strings.TrimSpace(req.GetSessionId()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "session_id is required")
	}
	if a.Learner == nil {
		return nil, onekiterrors.New(http.StatusServiceUnavailable, "feedback is unavailable")
	}
	rating := strings.ToLower(strings.TrimSpace(req.GetRating()))
	if rating != "up" && rating != "down" && rating != "clear" {
		return nil, onekiterrors.New(http.StatusBadRequest, "rating must be up, down or clear")
	}
	pending, err := a.Learner.RecordReplyFeedback(ctx, req.GetSessionId(), int(user.UserID), req.GetMessageId(), rating, req.GetReason())
	if err != nil {
		return nil, err
	}
	if a.AI != nil && application.ShouldDistillLessons(rating, pending) {
		go learnFromFeedback(context.WithoutCancel(ctx), a.AI, a.Learner, int(user.UserID))
	}
	return &hubcaretv1.RateReplyResponse{Saved: true}, nil
}

// learnFromFeedback runs one distillation pass; failures only delay learning.
func learnFromFeedback(ctx context.Context, ai application.TextCompleter, learner application.LessonLearner, userID int) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := application.DistillLessons(ctx, ai, learner, userID); err != nil {
		logger.Ctx(ctx).Warn("Caret could not learn from feedback", logger.Error(err))
	}
}

func (a *RuntimeAPI) ListTools(ctx context.Context, req *hubcaretv1.ListToolsRequest) (*hubcaretv1.ListToolsResponse, error) {
	resp, err := a.Tools.Serve(ctx, &ListToolsRequest{Mode: req.GetMode()})
	if err != nil {
		return nil, err
	}
	tools := make([]*hubcaretv1.CaretToolDef, len(resp.Tools))
	for i, tool := range resp.Tools {
		tools[i] = toolDefToProto(tool)
	}
	return &hubcaretv1.ListToolsResponse{Tools: tools}, nil
}

func (a *RuntimeAPI) GetProviderStatus(ctx context.Context, _ *hubcaretv1.GetProviderStatusRequest) (*hubcaretv1.GetProviderStatusResponse, error) {
	resp, err := a.Status.Serve(ctx, &GetProviderStatusRequest{})
	if err != nil {
		return nil, err
	}
	providers, err := json.Marshal(resp.Providers)
	if err != nil {
		return nil, err
	}
	return &hubcaretv1.GetProviderStatusResponse{
		Available:       resp.Available,
		PrimaryProvider: resp.PrimaryProvider,
		Providers:       providers,
	}, nil
}

func (a *RuntimeAPI) DecideTool(ctx context.Context, req *hubcaretv1.ToolDecisionRequest) (*hubcaretv1.ToolDecisionResponse, error) {
	var input map[string]any
	if raw := req.GetInput(); len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "input must be a JSON object")
		}
	}
	decision := &ToolDecisionRequest{
		ToolCallID: req.GetToolCallId(),
		Name:       req.GetName(),
		Input:      input,
		Decision:   req.GetDecision(),
		Mode:       req.GetMode(),
	}
	if err := a.Decision.Validate(ctx, decision); err != nil {
		return nil, err
	}
	resp, err := a.Decision.Serve(ctx, decision)
	if err != nil {
		return nil, err
	}
	var output json.RawMessage
	if resp.Output != nil {
		if output, err = json.Marshal(resp.Output); err != nil {
			return nil, err
		}
	}
	return &hubcaretv1.ToolDecisionResponse{
		ToolCallId: resp.ToolCallID,
		Status:     resp.Status,
		Message:    resp.Message,
		Executed:   resp.Executed,
		Tool:       toolDefToProto(resp.Tool),
		Preview:    resp.Preview,
		Output:     output,
		DurationMs: resp.DurationMs,
	}, nil
}

// StopReply ends the reply generating in one of the caller's chats.
func (a *RuntimeAPI) StopReply(ctx context.Context, req *hubcaretv1.StopReplyRequest) (*hubcaretv1.StopReplyResponse, error) {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 {
		return nil, onekiterrors.PermissionDenied()
	}
	if strings.TrimSpace(req.GetSessionId()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "session_id is required")
	}
	return &hubcaretv1.StopReplyResponse{Stopped: activeReplies.stop(user.UserID, req.GetSessionId())}, nil
}

func toolDefToProto(tool ToolDef) *hubcaretv1.CaretToolDef {
	return &hubcaretv1.CaretToolDef{
		Name:             tool.Name,
		Description:      tool.Description,
		Module:           tool.Module,
		Scope:            tool.Scope,
		RequiresApproval: tool.RequiresApproval,
		InputHint:        tool.InputHint,
		OutputHint:       tool.OutputHint,
		Freshness:        tool.Freshness,
		SideEffect:       tool.SideEffect,
		ApprovalReason:   tool.ApprovalReason,
	}
}
