package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/custom_domains/internal/application"
	"github.com/voxire/onion/internal/modules/custom_domains/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubconfigv1 "github.com/voxire/voxie/gen/go/hub/config/v1"
)

type API struct{ service *application.Service }

func New(service *application.Service) *API { return &API{service: service} }

func (a *API) GetCustomDomain(ctx context.Context, _ *hubconfigv1.GetCustomDomainRequest) (*hubconfigv1.GetCustomDomainResponse, error) {
	businessID, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := a.service.Get(ctx, businessID)
	if errors.Is(err, domain.ErrNotFound) {
		return &hubconfigv1.GetCustomDomainResponse{}, nil
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get custom domain")
	}
	return &hubconfigv1.GetCustomDomainResponse{Domain: payload(item, false)}, nil
}

func (a *API) RegisterCustomDomain(ctx context.Context, req *hubconfigv1.RegisterCustomDomainRequest) (*hubconfigv1.RegisterCustomDomainResponse, error) {
	device, businessID, err := deviceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := a.service.Register(ctx, device.User.UserID, businessID, req.GetDomain())
	if errors.Is(err, domain.ErrConflict) {
		return nil, onekiterrors.New(http.StatusConflict, err.Error())
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return &hubconfigv1.RegisterCustomDomainResponse{Domain: payload(item, true)}, nil
}

func (a *API) VerifyCustomDomain(ctx context.Context, req *hubconfigv1.VerifyCustomDomainRequest) (*hubconfigv1.VerifyCustomDomainResponse, error) {
	businessID, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	item, err := a.service.Verify(ctx, businessID, req.GetVerificationToken())
	if errors.Is(err, domain.ErrInvalidToken) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid verification token")
	}
	if errors.Is(err, domain.ErrNotFound) {
		return nil, onekiterrors.New(http.StatusNotFound, "Custom domain not found")
	}
	if errors.Is(err, domain.ErrDNSVerificationFailed) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Custom domain DNS record does not match")
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to verify custom domain")
	}
	return &hubconfigv1.VerifyCustomDomainResponse{Domain: payload(item, false)}, nil
}

func (a *API) RemoveCustomDomain(ctx context.Context, _ *hubconfigv1.RemoveCustomDomainRequest) (*hubconfigv1.RemoveCustomDomainResponse, error) {
	businessID, err := businessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.service.Remove(ctx, businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove custom domain")
	}
	return &hubconfigv1.RemoveCustomDomainResponse{Message: "Custom domain removed"}, nil
}

func businessID(ctx context.Context) (int64, error) {
	_, id, err := deviceBusinessID(ctx)
	return id, err
}

func deviceBusinessID(ctx context.Context) (authPkg.Device, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return authPkg.Device{}, 0, onekiterrors.PermissionDenied()
	}
	id, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return authPkg.Device{}, 0, onekiterrors.PermissionDenied()
	}
	return device, id, nil
}

func payload(item domain.CustomDomain, includeVerificationToken bool) *hubconfigv1.CustomDomain {
	var verified string
	if item.VerifiedAt != nil {
		verified = item.VerifiedAt.Format(time.RFC3339)
	}
	token := ""
	if includeVerificationToken {
		token = item.VerificationToken
	}
	return &hubconfigv1.CustomDomain{Id: item.ID, BusinessId: item.BusinessID, Domain: item.Domain, Status: item.Status, VerificationToken: token, SslStatus: item.SSLStatus, VerifiedAt: verified, CreatedAt: item.CreatedAt.Format(time.RFC3339), UpdatedAt: item.UpdatedAt.Format(time.RFC3339)}
}
