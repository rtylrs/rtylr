package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListFormFieldsRoutes() []string  { return []string{"/recruit/form-fields"} }
func CreateFormFieldRoutes() []string { return []string{"/recruit/form-fields"} }
func UpdateFormFieldRoutes() []string { return []string{"/recruit/form-fields"} }
func DeleteFormFieldRoutes() []string { return []string{"/recruit/form-fields"} }

// ─────────────────────────────────────────────
// Shared payload
// ─────────────────────────────────────────────

type FormFieldPayload struct {
	ID          int64  `json:"id"`
	CareerID    int64  `json:"career_id"`
	OwnerID     int64  `json:"owner_id"`
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
	OptionsJSON string `json:"options_json"`
	Position    int32  `json:"position"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toFormFieldPayload(f domain.FormField) *FormFieldPayload {
	return &FormFieldPayload{
		ID:          f.ID,
		CareerID:    f.CareerID,
		OwnerID:     f.OwnerID,
		Kind:        f.Kind,
		Label:       f.Label,
		Placeholder: f.Placeholder,
		Required:    f.Required,
		OptionsJSON: f.OptionsJSON,
		Position:    f.Position,
		CreatedAt:   f.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   f.UpdatedAt.Format(time.RFC3339),
	}
}

// ─────────────────────────────────────────────
// GET /recruit/form-fields?career_id=...
// ─────────────────────────────────────────────

type ListFormFieldsHandler struct {
	CareersService *application.Service
}

type ListFormFieldsRequest struct {
	CareerID int64 `json:"career_id"`
}

type ListFormFieldsResponse struct {
	Fields []*FormFieldPayload `json:"fields"`
}

func (h *ListFormFieldsHandler) Validate(_ context.Context, req *ListFormFieldsRequest) error {
	if req.CareerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "career_id is required")
	}
	return nil
}

func (h *ListFormFieldsHandler) Serve(ctx context.Context, req *ListFormFieldsRequest) (ListFormFieldsResponse, error) {
	authUser, permErr := requireRecruitPermission(ctx, "recruit:jobs:view")
	if permErr != nil {
		return ListFormFieldsResponse{}, permErr
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListFormFieldsResponse{}, onekiterrors.PermissionDenied()
	}
	fields, err := h.CareersService.ListFormFields(ctx, businessID, req.CareerID)
	if err != nil {
		return ListFormFieldsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list form fields")
	}
	payloads := make([]*FormFieldPayload, 0, len(fields))
	for _, f := range fields {
		payloads = append(payloads, toFormFieldPayload(f))
	}
	return ListFormFieldsResponse{Fields: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /recruit/form-fields
// ─────────────────────────────────────────────

type CreateFormFieldHandler struct {
	CareersService *application.Service
}

type CreateFormFieldRequest struct {
	CareerID    int64  `json:"career_id"`
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
	OptionsJSON string `json:"options_json"`
	Position    int32  `json:"position"`
}

type CreateFormFieldResponse struct {
	Field *FormFieldPayload `json:"field"`
}

func (h *CreateFormFieldHandler) Validate(_ context.Context, req *CreateFormFieldRequest) error {
	if req.CareerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "career_id is required")
	}
	if req.Kind == "" {
		return onekiterrors.New(http.StatusBadRequest, "kind is required")
	}
	if req.Label == "" {
		return onekiterrors.New(http.StatusBadRequest, "label is required")
	}
	return nil
}

func (h *CreateFormFieldHandler) Serve(ctx context.Context, req *CreateFormFieldRequest) (CreateFormFieldResponse, error) {
	authUser, permErr := requireRecruitPermission(ctx, "recruit:jobs:manage")
	if permErr != nil {
		return CreateFormFieldResponse{}, permErr
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateFormFieldResponse{}, onekiterrors.PermissionDenied()
	}

	in := domain.FormFieldInput{
		CareerID:    req.CareerID,
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Kind:        req.Kind,
		Label:       req.Label,
		Placeholder: req.Placeholder,
		Required:    req.Required,
		OptionsJSON: req.OptionsJSON,
		Position:    req.Position,
	}
	fieldID, err := h.CareersService.CreateFormField(ctx, in)
	if err != nil {
		return CreateFormFieldResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create field")
	}

	// Echo the row the insert produced. The list is ordered by position and
	// capped at 100 rows, so its tail is not necessarily the new field.
	field, err := h.CareersService.GetFormField(ctx, businessID, fieldID)
	if err != nil {
		return CreateFormFieldResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to load created field")
	}
	return CreateFormFieldResponse{Field: toFormFieldPayload(field)}, nil
}

// ─────────────────────────────────────────────
// PUT /recruit/form-fields
// ─────────────────────────────────────────────

type UpdateFormFieldHandler struct {
	CareersService *application.Service
}

type UpdateFormFieldRequest struct {
	ID          int64  `json:"id"`
	CareerID    int64  `json:"career_id"`
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
	OptionsJSON string `json:"options_json"`
	Position    int32  `json:"position"`
}

type UpdateFormFieldResponse struct {
	Message string `json:"message"`
}

func (h *UpdateFormFieldHandler) Validate(_ context.Context, req *UpdateFormFieldRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	if req.Label == "" {
		return onekiterrors.New(http.StatusBadRequest, "label is required")
	}
	return nil
}

func (h *UpdateFormFieldHandler) Serve(ctx context.Context, req *UpdateFormFieldRequest) (UpdateFormFieldResponse, error) {
	authUser, permErr := requireRecruitPermission(ctx, "recruit:jobs:manage")
	if permErr != nil {
		return UpdateFormFieldResponse{}, permErr
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateFormFieldResponse{}, onekiterrors.PermissionDenied()
	}

	in := domain.FormFieldInput{
		ID:          req.ID,
		CareerID:    req.CareerID,
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Kind:        req.Kind,
		Label:       req.Label,
		Placeholder: req.Placeholder,
		Required:    req.Required,
		OptionsJSON: req.OptionsJSON,
		Position:    req.Position,
	}
	if err := h.CareersService.UpdateFormField(ctx, in); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return UpdateFormFieldResponse{}, onekiterrors.New(http.StatusNotFound, "Field not found")
		}
		return UpdateFormFieldResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update field")
	}
	return UpdateFormFieldResponse{Message: "Field saved"}, nil
}

// ─────────────────────────────────────────────
// DELETE /recruit/form-fields
// ─────────────────────────────────────────────

type DeleteFormFieldHandler struct {
	CareersService *application.Service
}

type DeleteFormFieldRequest struct {
	ID int64 `json:"id"`
}

type DeleteFormFieldResponse struct {
	Message string `json:"message"`
}

func (h *DeleteFormFieldHandler) Validate(_ context.Context, req *DeleteFormFieldRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	return nil
}

func (h *DeleteFormFieldHandler) Serve(ctx context.Context, req *DeleteFormFieldRequest) (DeleteFormFieldResponse, error) {
	authUser, permErr := requireRecruitPermission(ctx, "recruit:jobs:manage")
	if permErr != nil {
		return DeleteFormFieldResponse{}, permErr
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteFormFieldResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.CareersService.DeleteFormField(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, application.ErrNotFound) {
			return DeleteFormFieldResponse{}, onekiterrors.New(http.StatusNotFound, "Field not found")
		}
		return DeleteFormFieldResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete field")
	}
	return DeleteFormFieldResponse{Message: "Field deleted"}, nil
}
