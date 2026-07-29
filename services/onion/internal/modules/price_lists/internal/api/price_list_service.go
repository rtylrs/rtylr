package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/price_lists/internal/application"
	"github.com/voxire/onion/internal/modules/price_lists/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erppricelistv1 "github.com/voxire/voxie/gen/go/erp/price_list/v1"
)

type PriceListAPI struct {
	Service *application.Service
}

func NewPriceListAPI(service *application.Service) *PriceListAPI {
	return &PriceListAPI{Service: service}
}

func (a *PriceListAPI) ListPriceLists(ctx context.Context, _ *erppricelistv1.ListPriceListsRequest) (*erppricelistv1.ListPriceListsResponse, error) {
	ownerID, err := priceListOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	lists, listErr := a.Service.ListPriceLists(ctx, ownerID)
	if listErr != nil {
		return nil, priceListStatusError(http.StatusInternalServerError, "Failed to list price lists")
	}
	out := make([]*erppricelistv1.PriceList, 0, len(lists))
	for _, p := range lists {
		out = append(out, newPriceListProto(p))
	}
	return &erppricelistv1.ListPriceListsResponse{PriceLists: out}, nil
}

func (a *PriceListAPI) CreatePriceList(ctx context.Context, req *erppricelistv1.CreatePriceListRequest) (*erppricelistv1.CreatePriceListResponse, error) {
	ownerID, err := priceListOwnerID(ctx)
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.CreatePriceList(ctx, ownerID, domain.PriceListInput{
		Name:      strings.TrimSpace(req.GetName()),
		Currency:  req.GetCurrency(),
		IsDefault: req.GetIsDefault(),
	})
	if createErr != nil {
		return nil, priceListStatusError(http.StatusInternalServerError, "Failed to create price list")
	}
	return &erppricelistv1.CreatePriceListResponse{
		PriceListId: id,
		Message:     "Price list created successfully",
	}, nil
}

func newPriceListProto(p domain.PriceList) *erppricelistv1.PriceList {
	return &erppricelistv1.PriceList{
		Id:        p.ID,
		OwnerId:   p.OwnerID,
		Name:      p.Name,
		Currency:  p.Currency,
		IsDefault: p.IsDefault,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func priceListOwnerID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, priceListStatusError(http.StatusUnauthorized, "unauthorized")
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return ownerID, nil
}

func priceListStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
