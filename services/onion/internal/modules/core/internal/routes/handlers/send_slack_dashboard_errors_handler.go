package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/core/internal/application"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func SlackDashboardSendRoutes() []string { return []string{"/slack/dashboard/send"} }

type SlackDashboardSendHandler struct {
	Service *application.Service
}

type SlackErrorLogRequest struct {
	Source         string   `json:"source"`
	Level          string   `json:"level"`
	Message        string   `json:"message"`
	Stack          string   `json:"stack"`
	ComponentStack string   `json:"componentStack"`
	URL            string   `json:"url"`
	Args           []string `json:"args"`
	MetaJSON       string   `json:"metaJson"`
	Formatted      string   `json:"formatted"`
}

type SlackErrorLogResponse struct {
	Status  string `json:"status"`
	Sent    bool   `json:"sent"`
	Message string `json:"message"`
}

func (h *SlackDashboardSendHandler) Validate(_ context.Context, r *SlackErrorLogRequest) *errorModels.ServerError {
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

	if r.URL != "" {
		parts = append(parts, "*URL:* "+r.URL)
	}

	if len(r.Args) > 0 {
		parts = append(parts, fmt.Sprintf("*Arguments:* %v", r.Args))
	}

	if r.MetaJSON != "" {
		parts = append(parts, fmt.Sprintf("*Metadata:* %v", r.MetaJSON))
	}

	r.Formatted = strings.Join(parts, "\n\n")

	return nil
}

func (h *SlackDashboardSendHandler) Serve(ctx context.Context, req *SlackErrorLogRequest) (SlackErrorLogResponse, *errorModels.ServerError) {
	h.Service.SendFormattedDashboardAlertsMessage(ctx, req.Formatted)

	return SlackErrorLogResponse{Status: "ok", Sent: true}, nil
}
