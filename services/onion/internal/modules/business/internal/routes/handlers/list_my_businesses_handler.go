package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/business/internal/application"
	"github.com/voxire/onion/internal/modules/business/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ListMyBusinessesRoutes() []string { return []string{"/businesses/mine"} }

type BusinessEntry struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	LogoURL  string `json:"logo_url"`
	ParentID int64  `json:"parent_id,omitempty"`
	IsActive bool   `json:"is_active"`
}

type (
	ListMyBusinessesHandler  struct{ Service *application.Service }
	ListMyBusinessesRequest  struct{}
	ListMyBusinessesResponse struct {
		Businesses []BusinessEntry `json:"businesses"`
	}
)

func toBusinessEntry(b domain.OwnedBusiness) BusinessEntry {
	return BusinessEntry{
		ID:       b.ID,
		Name:     b.Name,
		LogoURL:  b.LogoURL,
		ParentID: b.ParentID,
		IsActive: b.IsActive,
	}
}

func (h *ListMyBusinessesHandler) Serve(ctx context.Context, _ *ListMyBusinessesRequest) (ListMyBusinessesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListMyBusinessesResponse{}, errorModels.PermissionDeniedServerError()
	}

	businesses, err := h.Service.ListUserBusinesses(ctx, device.User.UserID)
	if err != nil {
		return ListMyBusinessesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to list businesses",
		}
	}

	entries := make([]BusinessEntry, 0, len(businesses))
	for _, b := range businesses {
		entries = append(entries, toBusinessEntry(b))
	}
	return ListMyBusinessesResponse{Businesses: entries}, nil
}
