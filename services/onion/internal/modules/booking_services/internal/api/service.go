package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking_services/internal/application"
	"github.com/voxire/onion/internal/modules/booking_services/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingservicev1 "github.com/voxire/voxie/gen/go/booking/service/v1"
)

type BookingServiceAPI struct {
	service *application.Service
}

func NewBookingServiceAPI(service *application.Service) *BookingServiceAPI {
	return &BookingServiceAPI{service: service}
}

func (a *BookingServiceAPI) ListServices(ctx context.Context, _ *bookingservicev1.ListServicesRequest) (*bookingservicev1.ListServicesResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list booking services")
	}
	return &bookingservicev1.ListServicesResponse{Services: newBookingServices(items)}, nil
}

func (a *BookingServiceAPI) CreateService(ctx context.Context, req *bookingservicev1.CreateServiceRequest) (*bookingservicev1.CreateServiceResponse, error) {
	input, errModel := newBookingServiceInput(req.GetName(), req.GetDescription(), req.GetDurationMinutes(), req.GetPrice(), req.GetColor(), req.GetIsActive())
	if errModel != nil {
		return nil, errModel
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, err := a.service.Create(ctx, businessID, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create booking service")
	}
	return &bookingservicev1.CreateServiceResponse{Id: id, Message: "Booking service created"}, nil
}

func (a *BookingServiceAPI) UpdateService(ctx context.Context, req *bookingservicev1.UpdateServiceRequest) (*bookingservicev1.UpdateServiceResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	input, errModel := newBookingServiceInput(req.GetName(), req.GetDescription(), req.GetDurationMinutes(), req.GetPrice(), req.GetColor(), req.GetIsActive())
	if errModel != nil {
		return nil, errModel
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.service.Update(ctx, businessID, req.GetId(), input); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update booking service")
	}
	return &bookingservicev1.UpdateServiceResponse{Message: "Booking service updated"}, nil
}

func (a *BookingServiceAPI) DeleteService(ctx context.Context, req *bookingservicev1.DeleteServiceRequest) (*bookingservicev1.DeleteServiceResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}

	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.service.Delete(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete booking service")
	}
	return &bookingservicev1.DeleteServiceResponse{Message: "Booking service deleted"}, nil
}

func newBookingServiceInput(name, description string, durationMinutes int32, price float64, color string, isActive bool) (domain.BookingServiceInput, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.BookingServiceInput{}, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if durationMinutes <= 0 {
		return domain.BookingServiceInput{}, onekiterrors.New(http.StatusBadRequest, "duration_minutes must be greater than 0")
	}
	amount, amountErr := money.FromFloat64(price)
	if amountErr != nil {
		return domain.BookingServiceInput{}, onekiterrors.New(http.StatusBadRequest, "price must be a valid monetary amount")
	}
	if amount < 0 {
		return domain.BookingServiceInput{}, onekiterrors.New(http.StatusBadRequest, "price must not be negative")
	}
	return domain.BookingServiceInput{
		Name:            name,
		Description:     description,
		DurationMinutes: int(durationMinutes),
		Price:           amount,
		Color:           color,
		IsActive:        isActive,
	}, nil
}

func newBookingServices(items []domain.BookingService) []*bookingservicev1.BookingService {
	out := make([]*bookingservicev1.BookingService, 0, len(items))
	for _, item := range items {
		out = append(out, &bookingservicev1.BookingService{
			Id:              item.ID,
			OwnerId:         item.OwnerID,
			BusinessId:      item.BusinessID,
			Name:            item.Name,
			Description:     item.Description,
			DurationMinutes: int32(item.DurationMinutes), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Price:           item.Price.Float64(),
			Color:           item.Color,
			IsActive:        item.IsActive,
			CreatedAt:       formatTime(item.CreatedAt),
			UpdatedAt:       formatTime(item.UpdatedAt),
		})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
