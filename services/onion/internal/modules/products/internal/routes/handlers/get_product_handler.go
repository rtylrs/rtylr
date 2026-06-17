package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

// GetProductHandlerFunc handles `GET /erp/products/{id}` — the
// FE-expected REST-path shape used by ProductDetail.tsx. The generic
// Register* helpers don't parse chi path params, so this one's wired
// manually in routes.go via router.Get. The OG PNG handler in the
// careers module uses the same pattern.

type GetProductResponse struct {
	Product *erpproductv1.Product `json:"product"`
}

func GetProductHandlerFunc(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authUser, ok := authPkg.GetAuthUser(ctx)
		if !ok {
			writeProductErr(ctx, w, http.StatusUnauthorized, "unauthorized")
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeProductErr(ctx, w, http.StatusBadRequest, "invalid product id")
			return
		}

		ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
		product, err := svc.GetProduct(ctx, ownerID, id)
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				writeProductErr(ctx, w, http.StatusNotFound, "product not found")
				return
			}
			logger.Ctx(ctx).Error("failed to load product",
				logger.Any("owner_id", ownerID),
				logger.Any("product_id", id),
				logger.Error(err),
			)
			writeProductErr(ctx, w, http.StatusInternalServerError, "failed to load product")
			return
		}

		resp := GetProductResponse{Product: newProductPayload(product)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
			logger.Ctx(ctx).Error("failed to encode product response", logger.Error(encErr))
		}
	}
}

func writeProductErr(ctx context.Context, w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := errorModels.ErrorResponse{
		Error: errorModels.ServerError{
			Code:     status,
			HTTPCode: status,
			Message:  msg,
		},
		Status: "error",
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Ctx(ctx).Error("failed to encode error response", logger.Error(err))
	}
}
