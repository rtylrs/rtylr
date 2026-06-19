package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/careers/internal/application"
	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListSectionsRoutes() []string    { return []string{"/recruit/sections"} }
func CreateSectionRoutes() []string   { return []string{"/recruit/sections"} }
func UpdateSectionRoutes() []string   { return []string{"/recruit/sections"} }
func DeleteSectionRoutes() []string   { return []string{"/recruit/sections"} }
func ReorderSectionsRoutes() []string { return []string{"/recruit/sections/reorder"} }

// ─────────────────────────────────────────────
// Shared payload
// ─────────────────────────────────────────────

type SectionPayload struct {
	ID        int64  `json:"id"`
	OwnerID   int64  `json:"owner_id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	MediaURL  string `json:"media_url"`
	ItemsJSON string `json:"items_json"`
	Position  int32  `json:"position"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toSectionPayload(s domain.Section) *SectionPayload {
	return &SectionPayload{
		ID:        s.ID,
		OwnerID:   s.OwnerID,
		Kind:      s.Kind,
		Title:     s.Title,
		Body:      s.Body,
		MediaURL:  s.MediaURL,
		ItemsJSON: s.ItemsJSON,
		Position:  s.Position,
		Enabled:   s.Enabled,
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	}
}

// ─────────────────────────────────────────────
// GET /recruit/sections — admin: all sections, including disabled
// ─────────────────────────────────────────────

type ListSectionsHandler struct {
	CareersService *application.Service
}

type ListSectionsRequest struct{}

type ListSectionsResponse struct {
	Sections []*SectionPayload `json:"sections"`
}

func (h *ListSectionsHandler) Serve(ctx context.Context, _ *ListSectionsRequest) (ListSectionsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListSectionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListSectionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	sections, err := h.CareersService.ListSections(ctx, businessID, false)
	if err != nil {
		return ListSectionsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list sections",
		}
	}
	payloads := make([]*SectionPayload, 0, len(sections))
	for _, s := range sections {
		payloads = append(payloads, toSectionPayload(s))
	}
	return ListSectionsResponse{Sections: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /recruit/sections
// ─────────────────────────────────────────────

type CreateSectionHandler struct {
	CareersService *application.Service
}

type CreateSectionRequest struct {
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	MediaURL  string `json:"media_url"`
	ItemsJSON string `json:"items_json"`
	Position  int32  `json:"position"`
	Enabled   bool   `json:"enabled"`
}

type CreateSectionResponse struct {
	Section *SectionPayload `json:"section"`
}

func (h *CreateSectionHandler) Validate(_ context.Context, req *CreateSectionRequest) *errorModels.ServerError {
	if req.Kind == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "kind is required"}
	}
	return nil
}

func (h *CreateSectionHandler) Serve(ctx context.Context, req *CreateSectionRequest) (CreateSectionResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateSectionResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateSectionResponse{}, errorModels.PermissionDeniedServerError()
	}

	in := domain.SectionInput{
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Kind:       req.Kind,
		Title:     req.Title,
		Body:      req.Body,
		MediaURL:  req.MediaURL,
		ItemsJSON: req.ItemsJSON,
		Position:  req.Position,
		Enabled:   req.Enabled,
	}

	_, err := h.CareersService.CreateSection(ctx, in)
	if err != nil {
		return CreateSectionResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create section",
		}
	}

	// Refetch list and return the newly-added row (tail of ordered
	// list when position wasn't supplied). Keeps the client in sync
	// without the client having to refetch.
	invalidatePublicCareersCache(ctx, businessID)

	sections, _ := h.CareersService.ListSections(ctx, businessID, false)
	if len(sections) == 0 {
		return CreateSectionResponse{Section: nil}, nil
	}
	return CreateSectionResponse{Section: toSectionPayload(sections[len(sections)-1])}, nil
}

// ─────────────────────────────────────────────
// PUT /recruit/sections
// ─────────────────────────────────────────────

type UpdateSectionHandler struct {
	CareersService *application.Service
}

type UpdateSectionRequest struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	MediaURL  string `json:"media_url"`
	ItemsJSON string `json:"items_json"`
	Position  int32  `json:"position"`
	Enabled   bool   `json:"enabled"`
}

type UpdateSectionResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSectionHandler) Validate(_ context.Context, req *UpdateSectionRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Kind == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "kind is required"}
	}
	return nil
}

func (h *UpdateSectionHandler) Serve(ctx context.Context, req *UpdateSectionRequest) (UpdateSectionResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateSectionResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateSectionResponse{}, errorModels.PermissionDeniedServerError()
	}

	in := domain.SectionInput{
		ID:         req.ID,
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Kind:       req.Kind,
		Title:     req.Title,
		Body:      req.Body,
		MediaURL:  req.MediaURL,
		ItemsJSON: req.ItemsJSON,
		Position:  req.Position,
		Enabled:   req.Enabled,
	}

	if err := h.CareersService.UpdateSection(ctx, in); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return UpdateSectionResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Section not found",
			}
		}
		return UpdateSectionResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update section",
		}
	}
	invalidatePublicCareersCache(ctx, businessID)
	return UpdateSectionResponse{Message: "Section saved"}, nil
}

// ─────────────────────────────────────────────
// DELETE /recruit/sections
// ─────────────────────────────────────────────

type DeleteSectionHandler struct {
	CareersService *application.Service
}

type DeleteSectionRequest struct {
	ID int64 `json:"id"`
}

type DeleteSectionResponse struct {
	Message string `json:"message"`
}

func (h *DeleteSectionHandler) Validate(_ context.Context, req *DeleteSectionRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteSectionHandler) Serve(ctx context.Context, req *DeleteSectionRequest) (DeleteSectionResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteSectionResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteSectionResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.CareersService.DeleteSection(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return DeleteSectionResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Section not found",
			}
		}
		return DeleteSectionResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete section",
		}
	}
	invalidatePublicCareersCache(ctx, businessID)
	return DeleteSectionResponse{Message: "Section deleted"}, nil
}

// ─────────────────────────────────────────────
// PUT /recruit/sections/reorder — batch position update
// ─────────────────────────────────────────────

type ReorderSectionsHandler struct {
	CareersService *application.Service
}

type ReorderSectionsRequest struct {
	OrderedIDs []int64 `json:"ordered_ids"`
}

type ReorderSectionsResponse struct {
	Message string `json:"message"`
}

func (h *ReorderSectionsHandler) Validate(_ context.Context, req *ReorderSectionsRequest) *errorModels.ServerError {
	if len(req.OrderedIDs) == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ordered_ids is required"}
	}
	return nil
}

func (h *ReorderSectionsHandler) Serve(ctx context.Context, req *ReorderSectionsRequest) (ReorderSectionsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ReorderSectionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ReorderSectionsResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.CareersService.ReorderSections(ctx, businessID, req.OrderedIDs); err != nil {
		return ReorderSectionsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to reorder sections",
		}
	}
	invalidatePublicCareersCache(ctx, businessID)
	return ReorderSectionsResponse{Message: "Order saved"}, nil
}
