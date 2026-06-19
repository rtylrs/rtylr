package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// FullProductResponse returns the complete product as JSON (every PIM column),
// unlike GetProduct which returns the proto payload that omits the newer fields.
// Used to fully prefill the edit form so a save round-trips without zeroing
// fields the proto can't carry yet.
type FullProductResponse struct {
	Product domain.Product `json:"product"`
}

// GetProductFullHandlerFunc handles `GET /erp/products/{id}/full`.
func GetProductFullHandlerFunc(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authUser, ok := authPkg.GetAuthUser(ctx)
		if !ok {
			writeProductErr(ctx, w, http.StatusUnauthorized, "unauthorized")
			return
		}

		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil || id <= 0 {
			writeProductErr(ctx, w, http.StatusBadRequest, "invalid product id")
			return
		}

		ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
		if bizErr != nil {
			writeProductErr(ctx, w, http.StatusForbidden, "forbidden")
			return
		}
		product, err := svc.GetProduct(ctx, ownerID, id)
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				writeProductErr(ctx, w, http.StatusNotFound, "product not found")
				return
			}
			logger.Ctx(ctx).Error("failed to load full product",
				logger.Any("product_id", id), logger.Error(err))
			writeProductErr(ctx, w, http.StatusInternalServerError, "failed to load product")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if encErr := json.NewEncoder(w).Encode(FullProductResponse{Product: product}); encErr != nil {
			logger.Ctx(ctx).Error("failed to encode full product response", logger.Error(encErr))
		}
	}
}
