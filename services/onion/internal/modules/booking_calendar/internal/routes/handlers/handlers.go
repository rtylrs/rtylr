package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/application"
	"github.com/voxire/onion/internal/modules/booking_calendar/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CalendarRoutes() []string { return []string{"/book/calendar"} }

type (
	CalendarHandler struct{ Service *application.Service }
	CalendarRequest struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
)

type CalendarResponse struct {
	Days []domain.CalendarDay `json:"days"`
}

func (h *CalendarHandler) Validate(_ context.Context, req *CalendarRequest) *errorModels.ServerError {
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	if req.From == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "from is required"}
	}
	if req.To == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "to is required"}
	}
	return nil
}

func (h *CalendarHandler) Serve(ctx context.Context, req *CalendarRequest) (CalendarResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CalendarResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CalendarResponse{}, errorModels.PermissionDeniedServerError()
	}
	days, err := h.Service.GetCalendar(ctx, businessID, req.From, req.To)
	if err != nil {
		return CalendarResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to get calendar"}
	}
	return CalendarResponse{Days: days}, nil
}
