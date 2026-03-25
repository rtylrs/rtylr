package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/core/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	corev1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/shared/core/v1"
)

func SlackDashboardSendRoutes() []string { return []string{"/slack/dashboard/send"} }

type SlackDashboardSendHandler struct {
	Service *application.Service
}

func (h *SlackDashboardSendHandler) Validate(_ context.Context, r *corev1.SlackErrorLogRequest) *errorModels.ServerError {
	if strings.TrimSpace(r.Source) == "" || strings.TrimSpace(r.Message) == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "source and message are required",
		}
	}

	allowedSources := map[string]struct{}{
		"console.error":      {},
		"window.error":       {},
		"unhandledrejection": {},
		"ErrorBoundary":      {},
	}
	if _, ok := allowedSources[r.Source]; !ok {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "invalid source",
		}
	}

	if r.Level == "" {
		r.Level = "error"
	}

	switch r.Level {
	case "error", "warn", "info":
	default:
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "invalid level",
		}
	}

	var parts []string

	level := strings.ToUpper(r.Level)

	parts = append(parts, fmt.Sprintf("*[%s] %s*", level, r.Source), r.Message)
	if r.Stack != "" {
		parts = append(parts, fmt.Sprintf("*Stack Trace:*\n```%s```", r.Stack))
	}

	if r.ComponentStack != "" {
		parts = append(parts, fmt.Sprintf("*Component Stack:*\n```%s```", r.ComponentStack))
	}

	if r.Url != "" {
		parts = append(parts, "*URL:* "+r.Url)
	}

	if len(r.Args) > 0 {
		parts = append(parts, fmt.Sprintf("*Arguments:* %v", r.Args))
	}

	if r.MetaJson != "" {
		parts = append(parts, fmt.Sprintf("*Metadata:* %v", r.MetaJson))
	}

	r.Formatted = strings.Join(parts, "\n\n")

	return nil
}

func (h *SlackDashboardSendHandler) Serve(ctx context.Context, req *corev1.SlackErrorLogRequest) (corev1.SlackErrorLogResponse, *errorModels.ServerError) {
	h.Service.SendFormattedDashboardAlertsMessage(ctx, req.Formatted)

	return corev1.SlackErrorLogResponse{Status: "ok", Sent: true}, nil
}
