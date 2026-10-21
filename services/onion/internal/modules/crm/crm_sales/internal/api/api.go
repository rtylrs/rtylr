package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// API implements the generated crm.sales.v1 services over the application
// service: it resolves the caller, checks their CRM permission, and
// translates between wire and domain types.
type API struct {
	Service *application.Service
}

func New(service *application.Service) *API {
	return &API{Service: service}
}

const (
	permView   = "crm:customers:view"
	permCreate = "crm:customers:create"
	permEdit   = "crm:customers:edit"
	permDelete = "crm:customers:delete"
)

type caller struct {
	businessID int64
	userID     int64
}

// caller resolves who is calling and checks they hold permission in CRM.
func (a *API) caller(ctx context.Context, permission string) (caller, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return caller{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, permission); err != nil {
		return caller{}, err
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return caller{}, onekiterrors.PermissionDenied()
	}
	return caller{businessID: businessID, userID: device.User.UserID}, nil
}

// fail turns a domain error into the API error the caller sees; anything
// unexpected is logged and reported as a failure of op.
func fail(ctx context.Context, op string, err error) error {
	var validation *domain.ValidationError
	switch {
	case errors.As(err, &validation):
		return onekiterrors.New(http.StatusBadRequest, validation.Message)
	case errors.Is(err, domain.ErrLeadNotFound):
		return onekiterrors.New(http.StatusNotFound, "Lead not found")
	case errors.Is(err, domain.ErrDealNotFound):
		return onekiterrors.New(http.StatusNotFound, "Deal not found")
	case errors.Is(err, domain.ErrPipelineNotFound):
		return onekiterrors.New(http.StatusNotFound, "Pipeline not found")
	case errors.Is(err, domain.ErrStageNotFound):
		return onekiterrors.New(http.StatusNotFound, "Stage not found")
	case errors.Is(err, domain.ErrTaskNotFound):
		return onekiterrors.New(http.StatusNotFound, "Task not found")
	case errors.Is(err, domain.ErrActivityNotFound):
		return onekiterrors.New(http.StatusNotFound, "Activity not found")
	case errors.Is(err, domain.ErrContractNotFound):
		return onekiterrors.New(http.StatusNotFound, "Contract not found")
	case errors.Is(err, domain.ErrImportNotFound):
		return onekiterrors.New(http.StatusNotFound, "Import not found")
	case errors.Is(err, domain.ErrInvalidAssignee):
		return onekiterrors.New(http.StatusBadRequest, "Assignee is not a member of this business")
	case errors.Is(err, domain.ErrInvalidCursor):
		return onekiterrors.New(http.StatusBadRequest, "Invalid cursor")
	case errors.Is(err, domain.ErrLeadConverted):
		return onekiterrors.New(http.StatusConflict, "This lead is converted; its work continues on its deal")
	case errors.Is(err, domain.ErrDealHasContract):
		return onekiterrors.New(http.StatusConflict, "This deal has a contract")
	case errors.Is(err, domain.ErrPipelineInUse):
		return onekiterrors.New(http.StatusConflict, "This pipeline still has deals")
	case errors.Is(err, domain.ErrStageInUse):
		return onekiterrors.New(http.StatusConflict, "A removed stage still has deals")
	case errors.Is(err, domain.ErrImportClosed):
		return onekiterrors.New(http.StatusConflict, "This import is no longer running")
	}
	logger.Ctx(ctx).Error("crm sales: "+op+" failed", logger.Error(err))
	return onekiterrors.New(http.StatusInternalServerError, "Failed to "+op)
}

func invalid(message string) error {
	return onekiterrors.New(http.StatusBadRequest, message)
}

func parseAmount(value, field string) (money.Amount, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	amount, err := money.FromString(value)
	if err != nil {
		return 0, invalid(field + " must be a decimal amount")
	}
	return amount, nil
}

func parseAmountPtr(value *string, field string) (*money.Amount, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	amount, err := parseAmount(*value, field)
	if err != nil {
		return nil, err
	}
	return &amount, nil
}

func amountPtrString(value *money.Amount) *string {
	if value == nil {
		return nil
	}
	s := value.String()
	return &s
}

func unixPtr(at *time.Time) *int64 {
	if at == nil {
		return nil
	}
	v := at.Unix()
	return &v
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func positivePtr(value *int64) *int64 {
	if value == nil || *value <= 0 {
		return nil
	}
	return value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

// splitList reads a comma-separated query parameter.
func splitList(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (a *API) callerDay(today string, utcOffsetMinutes int32) domain.CallerDay {
	return domain.NewCallerDay(today, utcOffsetMinutes, a.Service.Now())
}
