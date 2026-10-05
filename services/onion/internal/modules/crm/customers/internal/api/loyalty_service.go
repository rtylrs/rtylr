package api

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmloyaltyv1 "github.com/voxire/voxie/gen/go/crm/loyalty/v1"
)

type LoyaltyAPI struct {
	Service *application.Service
}

func NewLoyaltyAPI(service *application.Service) *LoyaltyAPI {
	return &LoyaltyAPI{Service: service}
}

func loyaltyBusinessID(ctx context.Context) (int64, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	return device.User.UserID, businessID, nil
}

// requireLoyaltyManager uses the same server-resolved permission as the chi
// handlers. The token's coarse role label is not authoritative for delegated
// CRM managers and can lag behind role/assignment changes.
func requireLoyaltyManager(ctx context.Context) error {
	return requireCRMFeaturePermission(ctx, "crm:loyalty:manage")
}

const (
	loyaltyTransactionsDefaultPageSize int64 = 20
	loyaltyTransactionsMaxPageSize     int64 = 100
)

func normalizeLoyaltyTransactionPage(page, pageSize int64) (int64, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = loyaltyTransactionsDefaultPageSize
	}
	if pageSize > loyaltyTransactionsMaxPageSize {
		pageSize = loyaltyTransactionsMaxPageSize
	}
	if page-1 > math.MaxInt32/pageSize {
		return 0, 0, onekiterrors.New(http.StatusBadRequest, "page offset is too large")
	}
	return page, pageSize, nil
}

func newLoyaltyConfigProto(cfg domain.LoyaltyConfig) *crmloyaltyv1.LoyaltyConfig {
	return &crmloyaltyv1.LoyaltyConfig{
		Id:                    cfg.ID,
		OwnerId:               cfg.OwnerID,
		IsEnabled:             cfg.IsEnabled,
		PointsPerCurrencyUnit: cfg.PointsPerCurrencyUnit,
		CurrencyPerPoint:      cfg.CurrencyPerPoint,
		BronzeMin:             int64(cfg.BronzeMin),
		SilverMin:             int64(cfg.SilverMin),
		GoldMin:               int64(cfg.GoldMin),
		PlatinumMin:           int64(cfg.PlatinumMin),
		BronzeDiscountPct:     cfg.BronzeDiscountPct,
		SilverDiscountPct:     cfg.SilverDiscountPct,
		GoldDiscountPct:       cfg.GoldDiscountPct,
		PlatinumDiscountPct:   cfg.PlatinumDiscountPct,
		PointsExpiryDays:      int64(cfg.PointsExpiryDays),
		UpdatedAt:             cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func newLoyaltyTierProto(t domain.LoyaltyTier) *crmloyaltyv1.LoyaltyTier {
	return &crmloyaltyv1.LoyaltyTier{
		Id:          t.ID,
		OwnerId:     t.OwnerID,
		Name:        t.Name,
		MinPoints:   int64(t.MinPoints),
		DiscountPct: t.DiscountPct,
		Perks:       string(t.Perks),
		SortOrder:   int64(t.SortOrder),
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func validateLoyaltyTierUpdateBody(body *crmloyaltyv1.UpdateLoyaltyTierBody) error {
	if body == nil {
		return onekiterrors.New(http.StatusBadRequest, "Loyalty tier body is required")
	}
	if err := body.Validate(); err != nil {
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	return nil
}

func newLoyaltyTransactionProto(tx domain.LoyaltyTransaction) *crmloyaltyv1.LoyaltyTransaction {
	return &crmloyaltyv1.LoyaltyTransaction{
		Id:            tx.ID,
		OwnerId:       tx.OwnerID,
		CustomerId:    tx.CustomerID,
		Type:          tx.Type,
		Points:        int64(tx.Points),
		BalanceBefore: int64(tx.BalanceBefore),
		BalanceAfter:  int64(tx.BalanceAfter),
		Description:   tx.Description,
		ReferenceId:   tx.ReferenceID,
		ReferenceType: tx.ReferenceType,
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (a *LoyaltyAPI) GetLoyaltyConfig(ctx context.Context, _ *crmloyaltyv1.GetLoyaltyConfigRequest) (*crmloyaltyv1.GetLoyaltyConfigResponse, error) {
	_, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	cfg, svcErr := a.Service.GetLoyaltyConfig(ctx, businessID)
	if svcErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get loyalty config")
	}

	return &crmloyaltyv1.GetLoyaltyConfigResponse{Config: newLoyaltyConfigProto(cfg)}, nil
}

func (a *LoyaltyAPI) UpdateLoyaltyConfig(ctx context.Context, req *crmloyaltyv1.UpdateLoyaltyConfigRequest) (*crmloyaltyv1.UpdateLoyaltyConfigResponse, error) {
	if err := requireLoyaltyManager(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Loyalty config is required")
	}
	if err := req.Validate(); err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	ownerID, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	// Same negative-value guards the HTTP handler enforces; a config with
	// negative earn rates or discount percentages corrupts every later
	// balance/tier calculation.
	switch {
	case req.GetPointsPerCurrencyUnit() < 0, req.GetCurrencyPerPoint() < 0,
		req.GetBronzeDiscountPct() < 0, req.GetSilverDiscountPct() < 0,
		req.GetGoldDiscountPct() < 0, req.GetPlatinumDiscountPct() < 0:
		return nil, onekiterrors.New(http.StatusBadRequest, "Loyalty config values must not be negative")
	}
	if err := validateLoyaltyConfigBounds(
		req.GetBronzeMin(),
		req.GetSilverMin(),
		req.GetGoldMin(),
		req.GetPlatinumMin(),
		req.GetPointsExpiryDays(),
	); err != nil {
		return nil, err
	}

	input := domain.LoyaltyConfigInput{
		OwnerID:               ownerID,
		BusinessID:            businessID,
		IsEnabled:             req.GetIsEnabled(),
		PointsPerCurrencyUnit: req.GetPointsPerCurrencyUnit(),
		CurrencyPerPoint:      req.GetCurrencyPerPoint(),
		BronzeMin:             int(req.GetBronzeMin()),
		SilverMin:             int(req.GetSilverMin()),
		GoldMin:               int(req.GetGoldMin()),
		PlatinumMin:           int(req.GetPlatinumMin()),
		BronzeDiscountPct:     req.GetBronzeDiscountPct(),
		SilverDiscountPct:     req.GetSilverDiscountPct(),
		GoldDiscountPct:       req.GetGoldDiscountPct(),
		PlatinumDiscountPct:   req.GetPlatinumDiscountPct(),
		PointsExpiryDays:      int(req.GetPointsExpiryDays()),
	}

	if svcErr := a.Service.UpsertLoyaltyConfig(ctx, input); svcErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update loyalty config")
	}

	return &crmloyaltyv1.UpdateLoyaltyConfigResponse{Message: "Loyalty config updated successfully"}, nil
}

func (a *LoyaltyAPI) ListLoyaltyTiers(ctx context.Context, _ *crmloyaltyv1.ListLoyaltyTiersRequest) (*crmloyaltyv1.ListLoyaltyTiersResponse, error) {
	_, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	tiers, svcErr := a.Service.ListLoyaltyTiers(ctx, businessID)
	if svcErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list loyalty tiers")
	}

	payloads := make([]*crmloyaltyv1.LoyaltyTier, len(tiers))
	for i, t := range tiers {
		payloads[i] = newLoyaltyTierProto(t)
	}

	return &crmloyaltyv1.ListLoyaltyTiersResponse{Tiers: payloads}, nil
}

func (a *LoyaltyAPI) CreateLoyaltyTier(ctx context.Context, req *crmloyaltyv1.CreateLoyaltyTierRequest) (*crmloyaltyv1.CreateLoyaltyTierResponse, error) {
	if err := requireLoyaltyManager(ctx); err != nil {
		return nil, err
	}

	ownerID, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	perks := []byte(req.GetPerks())
	if len(perks) == 0 {
		perks = []byte("[]")
	}

	id, svcErr := a.Service.CreateLoyaltyTier(ctx, domain.LoyaltyTierInput{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Name:        req.GetName(),
		MinPoints:   int(req.GetMinPoints()),
		DiscountPct: req.GetDiscountPct(),
		Perks:       perks,
		SortOrder:   int(req.GetSortOrder()),
	})
	if svcErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create loyalty tier")
	}

	return &crmloyaltyv1.CreateLoyaltyTierResponse{TierId: id, Message: "Loyalty tier created successfully"}, nil
}

func (a *LoyaltyAPI) UpdateLoyaltyTier(ctx context.Context, req *crmloyaltyv1.UpdateLoyaltyTierRequest) (*crmloyaltyv1.UpdateLoyaltyTierResponse, error) {
	if err := requireLoyaltyManager(ctx); err != nil {
		return nil, err
	}

	ownerID, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	body := req.GetBody()
	if err := validateLoyaltyTierUpdateBody(body); err != nil {
		return nil, err
	}
	perks := []byte(body.GetPerks())
	if len(perks) == 0 {
		perks = []byte("[]")
	}

	svcErr := a.Service.UpdateLoyaltyTier(ctx, domain.LoyaltyTierUpdateInput{
		ID:          req.GetId(),
		OwnerID:     ownerID,
		BusinessID:  businessID,
		Name:        body.GetName(),
		MinPoints:   int(body.GetMinPoints()),
		DiscountPct: body.GetDiscountPct(),
		Perks:       perks,
		SortOrder:   int(body.GetSortOrder()),
	})
	if svcErr != nil {
		if errors.Is(svcErr, domain.ErrTierNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Loyalty tier not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update loyalty tier")
	}

	return &crmloyaltyv1.UpdateLoyaltyTierResponse{Message: "Loyalty tier updated successfully"}, nil
}

func (a *LoyaltyAPI) ListLoyaltyTransactions(ctx context.Context, req *crmloyaltyv1.ListLoyaltyTransactionsRequest) (*crmloyaltyv1.ListLoyaltyTransactionsResponse, error) {
	_, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	var page, pageSize int64
	if req != nil {
		page, pageSize = req.GetPage(), req.GetPageSize()
	}
	page, pageSize, err = normalizeLoyaltyTransactionPage(page, pageSize)
	if err != nil {
		return nil, err
	}

	txs, total, svcErr := a.Service.ListLoyaltyTransactions(ctx, businessID, req.GetCustomerId(), int(page), int(pageSize))
	if svcErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list loyalty transactions")
	}

	payloads := make([]*crmloyaltyv1.LoyaltyTransaction, len(txs))
	for i, tx := range txs {
		payloads[i] = newLoyaltyTransactionProto(tx)
	}

	return &crmloyaltyv1.ListLoyaltyTransactionsResponse{
		Transactions: payloads,
		Total:        int64(total),
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

func (a *LoyaltyAPI) AwardPoints(ctx context.Context, req *crmloyaltyv1.AwardPointsRequest) (*crmloyaltyv1.AwardPointsResponse, error) {
	if err := requireLoyaltyManager(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Award points request is required")
	}
	ownerID, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetCustomerId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}
	if err := domain.ValidateLoyaltyPoints(req.GetPoints()); err != nil {
		if errors.Is(err, domain.ErrLoyaltyPointsOutOfRange) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Points exceed the maximum allowed value")
		}
		return nil, onekiterrors.New(http.StatusBadRequest, "Points must be greater than zero")
	}

	svcErr := a.Service.AwardPoints(ctx, domain.AwardPointsInput{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		CustomerID:    req.GetCustomerId(),
		Points:        int(req.GetPoints()),
		Description:   req.GetDescription(),
		ReferenceID:   req.ReferenceId,
		ReferenceType: req.GetReferenceType(),
	})
	if svcErr != nil {
		if errors.Is(svcErr, domain.ErrCustomerNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}
		if errors.Is(svcErr, domain.ErrLoyaltyBalanceCapacity) {
			return nil, onekiterrors.New(http.StatusConflict, "Loyalty balance capacity exceeded")
		}
		if errors.Is(svcErr, domain.ErrLoyaltyPointsOutOfRange) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Points exceed the maximum allowed value")
		}
		if errors.Is(svcErr, domain.ErrLoyaltyPointsNotPositive) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Points must be greater than zero")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to award points")
	}

	return &crmloyaltyv1.AwardPointsResponse{Message: "Points awarded successfully"}, nil
}

func (a *LoyaltyAPI) RedeemPoints(ctx context.Context, req *crmloyaltyv1.RedeemPointsRequest) (*crmloyaltyv1.RedeemPointsResponse, error) {
	if err := requireLoyaltyManager(ctx); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Redeem points request is required")
	}
	ownerID, businessID, err := loyaltyBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetCustomerId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}
	if err := domain.ValidateLoyaltyPoints(req.GetPoints()); err != nil {
		if errors.Is(err, domain.ErrLoyaltyPointsOutOfRange) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Points exceed the maximum allowed value")
		}
		return nil, onekiterrors.New(http.StatusBadRequest, "Points must be greater than zero")
	}

	svcErr := a.Service.RedeemPoints(ctx, domain.RedeemPointsInput{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		CustomerID:    req.GetCustomerId(),
		Points:        int(req.GetPoints()),
		Description:   req.GetDescription(),
		ReferenceID:   req.ReferenceId,
		ReferenceType: req.GetReferenceType(),
	})
	if svcErr != nil {
		if errors.Is(svcErr, domain.ErrCustomerNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}
		if errors.Is(svcErr, domain.ErrInsufficientPoints) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Insufficient loyalty points")
		}
		if errors.Is(svcErr, domain.ErrLoyaltyBalanceCapacity) {
			return nil, onekiterrors.New(http.StatusConflict, "Loyalty balance capacity exceeded")
		}
		if errors.Is(svcErr, domain.ErrLoyaltyPointsOutOfRange) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Points exceed the maximum allowed value")
		}
		if errors.Is(svcErr, domain.ErrLoyaltyPointsNotPositive) {
			return nil, onekiterrors.New(http.StatusBadRequest, "Points must be greater than zero")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to redeem points")
	}

	return &crmloyaltyv1.RedeemPointsResponse{Message: "Points redeemed successfully"}, nil
}
