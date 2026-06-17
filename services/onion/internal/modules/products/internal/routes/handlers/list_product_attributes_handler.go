package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ListProductAttributesResponse is the JSON body for the attributes endpoint.
type ListProductAttributesResponse struct {
	Attributes []domain.ProductAttribute `json:"attributes"`
}

// ListProductAttributesHandlerFunc handles `GET /erp/products/{id}/attributes` —
// the per-business custom attributes for a product, used to prefill the editor.
func ListProductAttributesHandlerFunc(svc *application.Service) http.HandlerFunc {
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

		ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
		attrs, err := svc.ListProductAttributes(ctx, ownerID, id)
		if err != nil {
			logger.Ctx(ctx).Error("failed to list product attributes",
				logger.Any("product_id", id), logger.Error(err))
			writeProductErr(ctx, w, http.StatusInternalServerError, "failed to load attributes")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if encErr := json.NewEncoder(w).Encode(ListProductAttributesResponse{Attributes: attrs}); encErr != nil {
			logger.Ctx(ctx).Error("failed to encode attributes response", logger.Error(encErr))
		}
	}
}
