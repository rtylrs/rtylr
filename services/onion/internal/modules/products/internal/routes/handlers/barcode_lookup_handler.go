package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

// BarcodeLookupResponse mirrors GetProductResponse — one product,
// nullable while the FE shows the lookup pending state.
type BarcodeLookupResponse struct {
	Product *erpproductv1.Product `json:"product"`
}

// BarcodeLookupHandlerFunc handles `GET /erp/products/barcode/{code}`.
// Same reason as GetProductHandlerFunc for being hand-wired — chi
// path params aren't parsed by the generic register helpers.
//
// Scope: owner-scoped lookup. A barcode that belongs to a different
// tenant returns 404, not 403, so the caller can't probe for cross-
// tenant barcode collisions.
func BarcodeLookupHandlerFunc(svc *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authUser, ok := authPkg.GetAuthUser(ctx)
		if !ok {
			writeProductErr(ctx, w, http.StatusUnauthorized, "unauthorized")
			return
		}

		code := chi.URLParam(r, "code")
		if code == "" {
			writeProductErr(ctx, w, http.StatusBadRequest, "barcode is required")
			return
		}

		ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
		product, err := svc.GetProductByBarcode(ctx, ownerID, code)
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				writeProductErr(ctx, w, http.StatusNotFound, "no product with that barcode")
				return
			}
			logger.Ctx(ctx).Error("barcode lookup failed",
				logger.Any("owner_id", ownerID),
				logger.String("barcode", code),
				logger.Error(err),
			)
			writeProductErr(ctx, w, http.StatusInternalServerError, "barcode lookup failed")
			return
		}

		resp := BarcodeLookupResponse{Product: newProductPayload(product)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
			logger.Ctx(ctx).Error("failed to encode barcode lookup response", logger.Error(encErr))
		}
	}
}
