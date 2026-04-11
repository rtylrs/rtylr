package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_services/internal/application"
	"github.com/voxire/onion/internal/modules/booking_services/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type Payload struct {
	ID int64 `json:"id"`; OwnerID int64 `json:"ownerId"`; Name string `json:"name"`
	Description string `json:"description"`; DurationMinutes int `json:"durationMinutes"`
	Price float64 `json:"price"`; Color string `json:"color"`; IsActive bool `json:"isActive"`
	CreatedAt string `json:"createdAt"`; UpdatedAt string `json:"updatedAt"`
}

func toPayload(s domain.BookingService) Payload {
	return Payload{ID: s.ID, OwnerID: s.OwnerID, Name: s.Name, Description: s.Description,
		DurationMinutes: s.DurationMinutes, Price: s.Price, Color: s.Color, IsActive: s.IsActive,
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func ListRoutes() []string   { return []string{"/book/services"} }
func CreateRoutes() []string { return []string{"/book/services"} }
func UpdateRoutes() []string { return []string{"/book/services"} }
func DeleteRoutes() []string { return []string{"/book/services"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Services []Payload `json:"services"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil { return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list booking services"} }
	out := make([]Payload, 0, len(items)); for _, i := range items { out = append(out, toPayload(i)) }
	return ListResponse{Services: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	Name string `json:"name"`; Description string `json:"description"`; DurationMinutes int `json:"durationMinutes"`
	Price float64 `json:"price"`; Color string `json:"color"`; IsActive bool `json:"isActive"`
}
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"} }; return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.BookingServiceInput{Name: req.Name, Description: req.Description, DurationMinutes: req.DurationMinutes, Price: req.Price, Color: req.Color, IsActive: req.IsActive})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create booking service"} }
	return CreateResponse{ID: id, Message: "Booking service created"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
	ID int64 `json:"id"`; Name string `json:"name"`; Description string `json:"description"`
	DurationMinutes int `json:"durationMinutes"`; Price float64 `json:"price"`; Color string `json:"color"`; IsActive bool `json:"isActive"`
}
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.BookingServiceInput{Name: req.Name, Description: req.Description, DurationMinutes: req.DurationMinutes, Price: req.Price, Color: req.Color, IsActive: req.IsActive}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update booking service"}
	}
	return UpdateResponse{Message: "Booking service updated"}, nil
}

type DeleteHandler struct{ Service *application.Service }
type DeleteRequest struct{ ID int64 `json:"id"` }
type DeleteResponse struct{ Message string `json:"message"` }

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Delete(ctx, authUser.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete booking service"}
	}
	return DeleteResponse{Message: "Booking service deleted"}, nil
}
