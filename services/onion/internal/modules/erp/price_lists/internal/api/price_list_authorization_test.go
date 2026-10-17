package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/price_lists/internal/application"
	"github.com/voxire/onion/internal/modules/erp/price_lists/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/voxie/gen/go/erp/price_list/v1"
)

type authorizationProbeRepository struct {
	listCalls   int
	createCalls int
}

func (r *authorizationProbeRepository) ListPriceLists(context.Context, int64) ([]domain.PriceList, error) {
	r.listCalls++
	return []domain.PriceList{}, nil
}

func (r *authorizationProbeRepository) CreatePriceList(context.Context, int64, domain.PriceListInput) (int64, error) {
	r.createCalls++
	return 1, nil
}

func TestPriceListAPIRequiresFeaturePermissions(t *testing.T) {
	repo := &authorizationProbeRepository{}
	api := NewPriceListAPI(application.NewService(context.Background(), repo))
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     12,
		BusinessID: 34,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID:     authorization.AppERP,
			Licensed:  true,
			HasAccess: true,
		}}},
	})

	if _, err := api.ListPriceLists(ctx, &v1.ListPriceListsRequest{}); err == nil {
		t.Fatal("ListPriceLists succeeded without erp:price_lists:view")
	}
	if _, err := api.CreatePriceList(ctx, &v1.CreatePriceListRequest{Name: "Unauthorized"}); err == nil {
		t.Fatal("CreatePriceList succeeded without erp:price_lists:manage")
	}
	if repo.listCalls != 0 || repo.createCalls != 0 {
		t.Fatalf("repository calls = list:%d create:%d, want neither operation to run", repo.listCalls, repo.createCalls)
	}
}
