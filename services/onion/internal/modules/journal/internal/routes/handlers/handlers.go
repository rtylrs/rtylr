package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/application"
	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type JournalLinePayload struct {
	ID        int64   `json:"id"`
	EntryID   int64   `json:"entryId"`
	AccountID int64   `json:"accountId"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

type JournalEntryPayload struct {
	ID          int64                 `json:"id"`
	OwnerID     int64                 `json:"ownerId"`
	Date        time.Time             `json:"date"`
	Description string                `json:"description"`
	Lines       []JournalLinePayload  `json:"lines"`
	CreatedAt   time.Time             `json:"createdAt"`
}

func newPayload(e domain.JournalEntry) JournalEntryPayload {
	lines := make([]JournalLinePayload, 0, len(e.Lines))
	for _, l := range e.Lines {
		lines = append(lines, JournalLinePayload{
			ID: l.ID, EntryID: l.EntryID, AccountID: l.AccountID,
			Debit: l.Debit, Credit: l.Credit,
		})
	}
	return JournalEntryPayload{
		ID: e.ID, OwnerID: e.OwnerID, Date: e.Date,
		Description: e.Description, Lines: lines, CreatedAt: e.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/finance/journal"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct {
	Entries []JournalEntryPayload `json:"entries"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list journal entries"}
	}
	payloads := make([]JournalEntryPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Entries: payloads}, nil
}

func CreateRoutes() []string { return []string{"/finance/journal"} }

type CreateHandler struct{ Service *application.Service }

type JournalLineInput struct {
	AccountID int64   `json:"accountId"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

type CreateRequest struct {
	Date        string             `json:"date"`
	Description string             `json:"description"`
	Lines       []JournalLineInput `json:"lines"`
}
type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Date == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "date is required"}
	}
	if len(req.Lines) == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "at least one line is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}

	lines := make([]domain.JournalLineInput, 0, len(req.Lines))
	for _, l := range req.Lines {
		lines = append(lines, domain.JournalLineInput{
			AccountID: l.AccountID, Debit: l.Debit, Credit: l.Credit,
		})
	}

	input := domain.JournalEntryInput{
		Date: req.Date, Description: req.Description, Lines: lines,
	}

	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create journal entry"}
	}

	return CreateResponse{ID: id, Message: "Journal entry created successfully"}, nil
}
