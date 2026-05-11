package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func ListTimeEntriesRoutes() []string { return []string{"/hr/time-entries"} }

type ListTimeEntriesHandler struct{ Service *application.Service }

type ListTimeEntriesRequest struct {
	EmployeeID  int64  `json:"employee_id"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	Limit       int32  `json:"limit"`
	Offset      int32  `json:"offset"`
}

type ListTimeEntriesResponse struct {
	Entries []*hrv1.TimeEntryPayload `json:"entries"`
	Total   int64                    `json:"total"`
	Limit   int32                    `json:"limit"`
	Offset  int32                    `json:"offset"`
}

func (h *ListTimeEntriesHandler) Serve(ctx context.Context, req *ListTimeEntriesRequest) (ListTimeEntriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListTimeEntriesResponse{}, errorModels.PermissionDeniedServerError()
	}

	entries, total, err := h.Service.ListTimeEntries(ctx, device.User.UserID, domain.ListTimeEntriesParams{
		EmployeeID:  req.EmployeeID,
		PeriodStart: req.PeriodStart,
		PeriodEnd:   req.PeriodEnd,
		Limit:       req.Limit,
		Offset:      req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list time entries", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListTimeEntriesResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list time entries"}
	}

	payloads := make([]*hrv1.TimeEntryPayload, 0, len(entries))
	for _, e := range entries {
		payloads = append(payloads, timeEntryToProto(e))
	}

	return ListTimeEntriesResponse{Entries: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
