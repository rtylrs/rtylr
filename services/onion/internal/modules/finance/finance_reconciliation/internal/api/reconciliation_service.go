package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/application"
	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financereconciliationv1 "github.com/voxire/voxie/gen/go/finance/reconciliation/v1"
)

type ReconciliationAPI struct {
	service *application.Service
}

func NewReconciliationAPI(service *application.Service) *ReconciliationAPI {
	return &ReconciliationAPI{service: service}
}

func (a *ReconciliationAPI) GetOverview(ctx context.Context, _ *financereconciliationv1.GetOverviewRequest) (*financereconciliationv1.GetOverviewResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.HasAppPermission(authUser, authorization.AppFinance, "finance:reconciliation:manage") {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	overview, err := a.service.Overview(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load reconciliation overview")
	}

	return newOverviewResponse(overview), nil
}

func (a *ReconciliationAPI) Match(ctx context.Context, req *financereconciliationv1.MatchRequest) (*financereconciliationv1.MatchResponse, error) {
	if req.GetBankFeedId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "bank_feed_id is required")
	}
	if req.GetTargetType() != domain.TargetTypeInvoice && req.GetTargetType() != domain.TargetTypeExpense {
		return nil, onekiterrors.New(http.StatusBadRequest, "target_type must be 'invoice' or 'expense'")
	}
	if req.GetTargetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "target_id is required")
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.HasAppPermission(authUser, authorization.AppFinance, "finance:reconciliation:manage") {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.service.Match(ctx, businessID, domain.MatchInput{
		BankFeedID: req.GetBankFeedId(),
		TargetType: req.GetTargetType(),
		TargetID:   req.GetTargetId(),
		MatchedBy:  authUser.UserID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrBankFeedNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Bank feed transaction not found")
		case errors.Is(err, domain.ErrTargetNotFound):
			return nil, onekiterrors.New(http.StatusNotFound, "Reconciliation target not found")
		case errors.Is(err, domain.ErrAlreadyMatched):
			return nil, onekiterrors.New(http.StatusConflict, "Bank feed transaction is already matched")
		case errors.Is(err, domain.ErrInvalidTarget):
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid match target")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create reconciliation match")
		}
	}

	return &financereconciliationv1.MatchResponse{Id: id, Message: "Reconciliation match recorded"}, nil
}

func newOverviewResponse(overview domain.ReconciliationOverview) *financereconciliationv1.GetOverviewResponse {
	bankTxns := make([]*financereconciliationv1.UnmatchedBankTransaction, 0, len(overview.UnmatchedBankTxns))
	for _, txn := range overview.UnmatchedBankTxns {
		bankTxns = append(bankTxns, &financereconciliationv1.UnmatchedBankTransaction{
			Id:               txn.ID,
			BankAccountId:    txn.BankAccountID,
			AmountMinorUnits: txn.AmountMinorUnits,
			Currency:         txn.Currency,
			OccurredAt:       formatTime(txn.OccurredAt),
			Description:      txn.Description,
			ExternalId:       txn.ExternalID,
		})
	}

	invoices := make([]*financereconciliationv1.UnmatchedInvoice, 0, len(overview.UnmatchedInvoices))
	for _, invoice := range overview.UnmatchedInvoices {
		invoices = append(invoices, &financereconciliationv1.UnmatchedInvoice{
			Id:               invoice.ID,
			InvoiceNumber:    invoice.InvoiceNumber,
			ClientName:       invoice.ClientName,
			AmountMinorUnits: invoice.AmountMinorUnits,
			Currency:         invoice.Currency,
			IssueDate:        formatTime(invoice.IssueDate),
			DueDate:          formatTime(invoice.DueDate),
		})
	}

	suggestions := make([]*financereconciliationv1.MatchSuggestion, 0, len(overview.Suggestions))
	for _, suggestion := range overview.Suggestions {
		suggestions = append(suggestions, &financereconciliationv1.MatchSuggestion{
			BankFeedId:       suggestion.BankFeedID,
			TargetType:       suggestion.TargetType,
			TargetId:         suggestion.TargetID,
			AmountMinorUnits: suggestion.AmountMinorUnits,
			DaysApart:        int32(suggestion.DaysApart), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Score:            suggestion.Score,
		})
	}

	return &financereconciliationv1.GetOverviewResponse{
		UnmatchedBankTxns: bankTxns,
		UnmatchedInvoices: invoices,
		Suggestions:       suggestions,
		BankTxnsTruncated: overview.BankTxnsTruncated,
		InvoicesTruncated: overview.InvoicesTruncated,
		ExpensesTruncated: overview.ExpensesTruncated,
		PreviewLimit:      int32(overview.PreviewLimit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
