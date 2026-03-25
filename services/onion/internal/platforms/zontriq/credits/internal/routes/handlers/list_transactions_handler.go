package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	commonv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/common/v1"
	creditsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/credits/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/application"
)

func ListTransactionsRoutes() []string {
	return []string{
		"/zontriq/credits/transactions",
	}
}

type ListTransactionsHandler struct {
	CreditsService *application.Service
}

func (h *ListTransactionsHandler) Serve(ctx context.Context, req *creditsv1.ListCreditTransactionsRequest) (creditsv1.ListCreditTransactionsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return creditsv1.ListCreditTransactionsResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	var (
		page     int32 = 1
		pageSize int32 = 20
	)

	if req.GetPagination() != nil {
		if req.GetPagination().GetPage() > 0 {
			page = req.GetPagination().GetPage()
		}

		if req.GetPagination().GetPageSize() > 0 {
			pageSize = req.GetPagination().GetPageSize()
		}
	}

	result, err := h.CreditsService.ListTransactions(ctx, application.ListTransactionsParams{
		UserID:   zontriqUserID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return creditsv1.ListCreditTransactionsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list transactions",
		}
	}

	protoTransactions := make([]*creditsv1.CreditTransaction, 0, len(result.Transactions))
	for i := range result.Transactions {
		t := &result.Transactions[i]
		protoTransactions = append(protoTransactions, &creditsv1.CreditTransaction{
			Id:            int64(t.ID),
			Amount:        t.Amount,
			Type:          t.Type,
			Description:   t.Description,
			ReferenceType: t.ReferenceType,
			ReferenceId:   int64(t.ReferenceID),
			Status:        t.Status,
			CreatedAt:     t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	totalPages := int32(result.TotalCount) / pageSize
	if int32(result.TotalCount)%pageSize > 0 {
		totalPages++
	}

	return creditsv1.ListCreditTransactionsResponse{
		Transactions: protoTransactions,
		Pagination: &commonv1.PaginationResponse{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
			TotalCount:  result.TotalCount,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}
