package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/journal/internal/application"
	"github.com/voxire/onion/internal/modules/journal/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	commonpaginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
	financejournalv1 "github.com/voxire/voxie/gen/go/finance/journal/v1"
)

const journalDateLayout = "2006-01-02"

type JournalAPI struct {
	Service *application.Service
}

func NewJournalAPI(service *application.Service) *JournalAPI {
	return &JournalAPI{Service: service}
}

func (a *JournalAPI) ListJournalEntries(ctx context.Context, req *financejournalv1.ListJournalEntriesRequest) (*financejournalv1.ListJournalEntriesResponse, error) {
	businessID, err := journalBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	limit := req.GetLimit()
	if limit < 1 {
		limit = 50
	}
	entries, hasMore, listErr := a.Service.List(ctx, businessID, int64(page), int64(limit))
	if listErr != nil {
		return nil, journalStatusError(http.StatusInternalServerError, "Failed to list journal entries")
	}
	payloads := make([]*financejournalv1.JournalEntry, 0, len(entries))
	for _, entry := range entries {
		payloads = append(payloads, newJournalEntryPayload(entry))
	}
	return &financejournalv1.ListJournalEntriesResponse{
		Entries: payloads,
		Pagination: &commonpaginationv1.PaginationResponse{
			Page:    page,
			Limit:   limit,
			HasNext: hasMore,
			HasPrev: page > 1,
		},
	}, nil
}

func (a *JournalAPI) CreateJournalEntry(ctx context.Context, req *financejournalv1.CreateJournalEntryRequest) (*financejournalv1.CreateJournalEntryResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	date := strings.TrimSpace(req.GetDate())
	if date == "" {
		return nil, journalStatusError(http.StatusBadRequest, "date is required")
	}
	if _, err := time.Parse(journalDateLayout, date); err != nil {
		return nil, journalStatusError(http.StatusBadRequest, "date must be in YYYY-MM-DD format")
	}
	if len(req.GetLines()) == 0 {
		return nil, journalStatusError(http.StatusBadRequest, "at least one line is required")
	}
	lines := make([]domain.JournalLineInput, 0, len(req.GetLines()))
	for _, line := range req.GetLines() {
		lines = append(lines, domain.JournalLineInput{
			AccountID: line.GetAccountId(),
			Debit:     line.GetDebit(),
			Credit:    line.GetCredit(),
		})
	}
	id, createErr := a.Service.Create(ctx, device.User.UserID, businessID, domain.JournalEntryInput{
		Date:        date,
		Description: req.GetDescription(),
		Lines:       lines,
	})
	if createErr != nil {
		if errors.Is(createErr, domain.ErrInvalidJournalEntry) {
			return nil, journalStatusError(http.StatusBadRequest, createErr.Error())
		}
		return nil, journalStatusError(http.StatusInternalServerError, "Failed to create journal entry")
	}
	return &financejournalv1.CreateJournalEntryResponse{
		Entry: &financejournalv1.JournalEntry{
			Id:          id,
			Date:        date,
			Description: req.GetDescription(),
			Lines:       req.GetLines(),
			Status:      "draft",
		},
	}, nil
}

func (a *JournalAPI) GetJournalEntry(ctx context.Context, req *financejournalv1.GetJournalEntryRequest) (*financejournalv1.GetJournalEntryResponse, error) {
	businessID, err := journalBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	entry, getErr := a.Service.Get(ctx, businessID, req.GetId())
	if errors.Is(getErr, domain.ErrJournalEntryNotFound) {
		return nil, journalStatusError(http.StatusNotFound, "Journal entry not found")
	}
	if getErr != nil {
		return nil, journalStatusError(http.StatusInternalServerError, "Failed to get journal entry")
	}
	return &financejournalv1.GetJournalEntryResponse{Entry: newJournalEntryPayload(entry)}, nil
}

func newJournalEntryPayload(entry domain.JournalEntry) *financejournalv1.JournalEntry {
	lines := make([]*financejournalv1.JournalLine, 0, len(entry.Lines))
	for _, line := range entry.Lines {
		lines = append(lines, &financejournalv1.JournalLine{
			AccountId:   line.AccountID,
			AccountName: line.AccountName,
			Debit:       line.Debit,
			Credit:      line.Credit,
		})
	}
	return &financejournalv1.JournalEntry{
		Id:          entry.ID,
		Date:        entry.Date.Format(journalDateLayout),
		Description: entry.Description,
		Lines:       lines,
		Status:      entry.Status,
		CreatedAt:   entry.CreatedAt.Format(time.RFC3339),
	}
}

func journalBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func journalStatusError(code int, message string) error {
	return onekiterrors.New(code, message)
}
