package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── Delete Product by Path ID ────────────────────────────────────────────────
//
// DELETE /erp/products/{id} — REST-style variant. Coexists with the existing
// flat DELETE /erp/products (body-ID) route. Reuses Service.DeleteProduct.

func DeleteProductByIDRoutes() []string {
	return []string{"/erp/products/{id}"}
}

type DeleteProductByIDHandler struct {
	Service *application.Service
}

type DeleteProductByIDRequest struct{}

type DeleteProductByIDResponse struct {
	Message string `json:"message"`
}

func (h *DeleteProductByIDHandler) Serve(ctx context.Context, _ *DeleteProductByIDRequest) (DeleteProductByIDResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteProductByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	idRaw := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		return DeleteProductByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid product id",
		}
	}

	if err := h.Service.DeleteProduct(ctx, authUser.UserID, id); err != nil {
		return DeleteProductByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete product",
		}
	}

	return DeleteProductByIDResponse{Message: "Product deleted successfully"}, nil
}
