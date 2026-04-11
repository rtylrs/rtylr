package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/application"
	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type Payload struct {
	ID int64 `json:"id"`; OwnerID int64 `json:"ownerId"`; ServiceID int64 `json:"serviceId"`
	CustomerName string `json:"customerName"`; CustomerEmail string `json:"customerEmail"`; CustomerPhone string `json:"customerPhone"`
	PreferredDate string `json:"preferredDate"`; PreferredTime string `json:"preferredTime"`; Status string `json:"status"`; CreatedAt string `json:"createdAt"`
}

func toPayload(w domain.WaitlistEntry) Payload {
	return Payload{ID: w.ID, OwnerID: w.OwnerID, ServiceID: w.ServiceID, CustomerName: w.CustomerName,
		CustomerEmail: w.CustomerEmail, CustomerPhone: w.CustomerPhone, PreferredDate: w.PreferredDate,
		PreferredTime: w.PreferredTime, Status: w.Status, CreatedAt: w.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func ListRoutes() []string   { return []string{"/book/waitlist"} }
func CreateRoutes() []string { return []string{"/book/waitlist"} }
func UpdateRoutes() []string { return []string{"/book/waitlist"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Entries []Payload `json:"entries"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil { return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list waitlist"} }
	out := make([]Payload, 0, len(items)); for _, i := range items { out = append(out, toPayload(i)) }
	return ListResponse{Entries: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	ServiceID int64 `json:"serviceId"`; CustomerName string `json:"customerName"`; CustomerEmail string `json:"customerEmail"`
	CustomerPhone string `json:"customerPhone"`; PreferredDate string `json:"preferredDate"`; PreferredTime string `json:"preferredTime"`
}
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, _ *CreateRequest) *errorModels.ServerError { return nil }

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.WaitlistInput{
		ServiceID: req.ServiceID, CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone, PreferredDate: req.PreferredDate, PreferredTime: req.PreferredTime,
	})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create waitlist entry"} }
	return CreateResponse{ID: id, Message: "Waitlist entry created"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
	ID int64 `json:"id"`; ServiceID int64 `json:"serviceId"`; CustomerName string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`; CustomerPhone string `json:"customerPhone"`
	PreferredDate string `json:"preferredDate"`; PreferredTime string `json:"preferredTime"`; Status string `json:"status"`
}
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.WaitlistInput{
		ServiceID: req.ServiceID, CustomerName: req.CustomerName, CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone, PreferredDate: req.PreferredDate, PreferredTime: req.PreferredTime, Status: req.Status,
	}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update waitlist entry"}
	}
	return UpdateResponse{Message: "Waitlist entry updated"}, nil
}
