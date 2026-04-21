package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/voxire/onion/internal/modules/journal/internal/application"
	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type JournalLinePayload struct {
	ID        int64   `json:"id"`
	EntryID   int64   `json:"entry_id"`
	AccountID int64   `json:"account_id"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

type JournalEntryPayload struct {
	ID          int64                `json:"id"`
	OwnerID     int64                `json:"owner_id"`
	Date        time.Time            `json:"date"`
	Description string               `json:"description"`
	Status      string               `json:"status"`
	PostedAt    *time.Time           `json:"posted_at"`
	Lines       []JournalLinePayload `json:"lines"`
	CreatedAt   time.Time            `json:"created_at"`
}

func newPayload(e domain.JournalEntry) JournalEntryPayload {
	lines := make([]JournalLinePayload, 0, len(e.Lines))
	for _, l := range e.Lines {
		lines = append(lines, JournalLinePayload{
			ID: l.ID, EntryID: l.EntryID, AccountID: l.AccountID,
			Debit: l.Debit, Credit: l.Credit,
		})
	}
	var postedAt *time.Time
	if !e.PostedAt.IsZero() {
		pt := e.PostedAt
		postedAt = &pt
	}
	return JournalEntryPayload{
		ID: e.ID, OwnerID: e.OwnerID, Date: e.Date,
		Description: e.Description, Status: e.Status, PostedAt: postedAt,
		Lines: lines, CreatedAt: e.CreatedAt,
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
	AccountID int64   `json:"account_id"`
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

// ---------------------------------------------------------------------------
// Post Journal Entry
// ---------------------------------------------------------------------------

func PostRoutes() []string { return []string{"/finance/journal/{id}/post"} }

type PostHandler struct{ Service *application.Service }
type PostRequest struct{}
type PostResponse struct {
	ID      int64  `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (h *PostHandler) Serve(ctx context.Context, _ *PostRequest) (PostResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return PostResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return PostResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid journal entry ID"}
	}

	if postErr := h.Service.Post(ctx, device.User.UserID, id); postErr != nil {
		switch {
		case errors.Is(postErr, domain.ErrJournalEntryNotFound):
			return PostResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Journal entry not found"}
		case errors.Is(postErr, domain.ErrJournalEntryAlreadyPosted):
			return PostResponse{}, &errorModels.ServerError{Code: http.StatusConflict, Message: "Journal entry already posted"}
		default:
			return PostResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to post journal entry"}
		}
	}

	return PostResponse{ID: id, Status: "posted", Message: "Journal entry posted successfully"}, nil
}
