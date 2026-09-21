package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/application"
	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	bookingwaitlistv1 "github.com/voxire/voxie/gen/go/booking/waitlist/v1"
)

type WaitlistAPI struct {
	service *application.Service
}

func NewWaitlistAPI(service *application.Service) *WaitlistAPI {
	return &WaitlistAPI{service: service}
}

func (a *WaitlistAPI) ListEntries(ctx context.Context, _ *bookingwaitlistv1.ListEntriesRequest) (*bookingwaitlistv1.ListEntriesResponse, error) {
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
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list waitlist")
	}
	return &bookingwaitlistv1.ListEntriesResponse{Entries: newWaitlistEntries(items)}, nil
}

func (a *WaitlistAPI) CreateEntry(ctx context.Context, req *bookingwaitlistv1.CreateEntryRequest) (*bookingwaitlistv1.CreateEntryResponse, error) {
	input, errModel := newWaitlistInput(req.GetServiceId(), req.GetCustomerName(), req.GetCustomerEmail(), req.GetCustomerPhone(), req.GetPreferredDate(), req.GetPreferredTime(), "", true)
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
		if errors.Is(err, domain.ErrServiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Booking service not found")
		}
		if errors.Is(err, domain.ErrPreferredDateInvalid) || errors.Is(err, domain.ErrPreferredTimeInvalid) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create waitlist entry")
	}
	return &bookingwaitlistv1.CreateEntryResponse{Id: id, Message: "Waitlist entry created"}, nil
}

func (a *WaitlistAPI) UpdateEntry(ctx context.Context, req *bookingwaitlistv1.UpdateEntryRequest) (*bookingwaitlistv1.UpdateEntryResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	input, errModel := newWaitlistInput(req.GetServiceId(), req.GetCustomerName(), req.GetCustomerEmail(), req.GetCustomerPhone(), req.GetPreferredDate(), req.GetPreferredTime(), req.GetStatus(), false)
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
		if errors.Is(err, domain.ErrServiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Booking service not found")
		}
		if errors.Is(err, domain.ErrWaitlistEntryNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Waitlist entry not found")
		}
		if errors.Is(err, domain.ErrPreferredDateInvalid) || errors.Is(err, domain.ErrPreferredTimeInvalid) {
			return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update waitlist entry")
	}
	return &bookingwaitlistv1.UpdateEntryResponse{Message: "Waitlist entry updated"}, nil
}

func newWaitlistInput(serviceID int64, customerName, customerEmail, customerPhone, preferredDate, preferredTime, status string, requireName bool) (domain.WaitlistInput, error) {
	customerName = strings.TrimSpace(customerName)
	if requireName && customerName == "" {
		return domain.WaitlistInput{}, onekiterrors.New(http.StatusBadRequest, "customer_name is required")
	}
	return domain.WaitlistInput{
		ServiceID:     serviceID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		CustomerPhone: customerPhone,
		PreferredDate: preferredDate,
		PreferredTime: preferredTime,
		Status:        status,
	}, nil
}

func newWaitlistEntries(items []domain.WaitlistEntry) []*bookingwaitlistv1.WaitlistEntry {
	out := make([]*bookingwaitlistv1.WaitlistEntry, 0, len(items))
	for _, item := range items {
		out = append(out, &bookingwaitlistv1.WaitlistEntry{
			Id:            item.ID,
			OwnerId:       item.OwnerID,
			BusinessId:    item.BusinessID,
			ServiceId:     item.ServiceID,
			CustomerName:  item.CustomerName,
			CustomerEmail: item.CustomerEmail,
			CustomerPhone: item.CustomerPhone,
			PreferredDate: item.PreferredDate,
			PreferredTime: item.PreferredTime,
			Status:        item.Status,
			CreatedAt:     formatTime(item.CreatedAt),
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
