package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/settings/scheduled-reports

func ListScheduledReportsRoutes() []string { return []string{"/insights/settings/scheduled-reports"} }

type ListScheduledReportsHandler struct{ Service *application.Service }

type ListScheduledReportsRequest struct{}

type ListScheduledReportsResponse struct {
	ScheduledReports []application.ScheduledReportPayload `json:"scheduled_reports"`
}

func (h *ListScheduledReportsHandler) Serve(ctx context.Context, _ *ListScheduledReportsRequest) (ListScheduledReportsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListScheduledReportsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListScheduledReportsResponse{}, errorModels.PermissionDeniedServerError()
	}
	rows, err := h.Service.ListScheduledReports(ctx, businessID)
	if err != nil {
		return ListScheduledReportsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list scheduled reports",
		}
	}
	return ListScheduledReportsResponse{ScheduledReports: rows}, nil
}

// POST /insights/settings/scheduled-reports

func CreateScheduledReportRoutes() []string { return []string{"/insights/settings/scheduled-reports"} }

type CreateScheduledReportHandler struct{ Service *application.Service }

type CreateScheduledReportRequest struct {
	TemplateID string `json:"template_id"`
	Cadence    string `json:"cadence"`
	Channel    string `json:"channel"`
	Recipient  string `json:"recipient"`
	IsActive   bool   `json:"is_active"`
}

type CreateScheduledReportResponse = application.ScheduledReportPayload

func (h *CreateScheduledReportHandler) Validate(_ context.Context, req *CreateScheduledReportRequest) *errorModels.ServerError {
	req.TemplateID = strings.TrimSpace(req.TemplateID)
	if req.TemplateID == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "template_id is required",
		}
	}
	req.Cadence = strings.TrimSpace(req.Cadence)
	if req.Cadence == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "cadence is required",
		}
	}
	req.Recipient = strings.TrimSpace(req.Recipient)
	if req.Recipient == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "recipient is required",
		}
	}
	return nil
}

func (h *CreateScheduledReportHandler) Serve(ctx context.Context, req *CreateScheduledReportRequest) (CreateScheduledReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateScheduledReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateScheduledReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	out, err := h.Service.CreateScheduledReport(ctx, businessID, application.ScheduledReportPayload{
		TemplateID: req.TemplateID,
		Cadence:    req.Cadence,
		Channel:    req.Channel,
		Recipient:  req.Recipient,
		IsActive:   req.IsActive,
	})
	if err != nil {
		return CreateScheduledReportResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  err.Error(),
		}
	}
	return out, nil
}

// PUT /insights/settings/scheduled-reports/{id}

func UpdateScheduledReportRoutes() []string {
	return []string{"/insights/settings/scheduled-reports/{id}"}
}

type UpdateScheduledReportHandler struct{ Service *application.Service }

type UpdateScheduledReportRequest struct {
	TemplateID string `json:"template_id"`
	Cadence    string `json:"cadence"`
	Channel    string `json:"channel"`
	Recipient  string `json:"recipient"`
	IsActive   bool   `json:"is_active"`
}

type UpdateScheduledReportResponse = application.ScheduledReportPayload

func (h *UpdateScheduledReportHandler) Validate(_ context.Context, req *UpdateScheduledReportRequest) *errorModels.ServerError {
	req.TemplateID = strings.TrimSpace(req.TemplateID)
	if req.TemplateID == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "template_id is required",
		}
	}
	req.Cadence = strings.TrimSpace(req.Cadence)
	if req.Cadence == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "cadence is required",
		}
	}
	req.Recipient = strings.TrimSpace(req.Recipient)
	if req.Recipient == "" {
		return &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "recipient is required",
		}
	}
	return nil
}

func (h *UpdateScheduledReportHandler) Serve(ctx context.Context, req *UpdateScheduledReportRequest) (UpdateScheduledReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateScheduledReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateScheduledReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := parseIDParam(ctx)
	if err != nil {
		return UpdateScheduledReportResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid id",
		}
	}
	out, err := h.Service.UpdateScheduledReport(ctx, businessID, id, application.ScheduledReportPayload{
		TemplateID: req.TemplateID,
		Cadence:    req.Cadence,
		Channel:    req.Channel,
		Recipient:  req.Recipient,
		IsActive:   req.IsActive,
	})
	if err != nil {
		if errors.Is(err, application.ScheduledReportErrNotFound) {
			return UpdateScheduledReportResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Scheduled report not found",
			}
		}
		return UpdateScheduledReportResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  err.Error(),
		}
	}
	return out, nil
}

// DELETE /insights/settings/scheduled-reports/{id}

func DeleteScheduledReportRoutes() []string {
	return []string{"/insights/settings/scheduled-reports/{id}"}
}

type DeleteScheduledReportHandler struct{ Service *application.Service }

type DeleteScheduledReportRequest struct{}

type DeleteScheduledReportResponse struct {
	Deleted bool `json:"deleted"`
}

func (h *DeleteScheduledReportHandler) Serve(ctx context.Context, _ *DeleteScheduledReportRequest) (DeleteScheduledReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteScheduledReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteScheduledReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := parseIDParam(ctx)
	if err != nil {
		return DeleteScheduledReportResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid id",
		}
	}
	if delErr := h.Service.DeleteScheduledReport(ctx, businessID, id); delErr != nil {
		if errors.Is(delErr, application.ScheduledReportErrNotFound) {
			return DeleteScheduledReportResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusNotFound,
				Code:     http.StatusNotFound,
				Message:  "Scheduled report not found",
			}
		}
		return DeleteScheduledReportResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to delete scheduled report",
		}
	}
	return DeleteScheduledReportResponse{Deleted: true}, nil
}

// parseIDParam lifts the {id} chi URL param into an int64 with validation.
func parseIDParam(ctx context.Context) (int64, error) {
	raw := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}
