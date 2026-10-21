package api

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/products/internal/application"
	"github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

func newAuthorizationTestProductService(ctx context.Context) *application.Service {
	return application.NewService(infrastructure.NewRepository(testutil.DirectAuthorizationDB(ctx)))
}

func TestDirectAuthorizationMatrix(t *testing.T) {
	cases := []testutil.DirectAuthorizationCase{
		{Name: "erp products view", App: authorization.AppERP, Permission: "erp:products:view", Call: func(ctx context.Context) (any, error) {
			api := NewProductAPI(newAuthorizationTestProductService(ctx))
			response, err := api.ListProducts(ctx, &erpproductv1.ListProductsRequest{})
			return response, err
		}},
		{Name: "erp products create", App: authorization.AppERP, Permission: "erp:products:create", Call: func(ctx context.Context) (any, error) {
			api := NewProductAPI(newAuthorizationTestProductService(ctx))
			response, err := api.CreateProduct(ctx, &erpproductv1.CreateProductRequest{Name: "authorization probe"})
			return response, err
		}},
		{Name: "erp products edit", App: authorization.AppERP, Permission: "erp:products:edit", Call: func(ctx context.Context) (any, error) {
			api := NewProductAPI(newAuthorizationTestProductService(ctx))
			response, err := api.UpdateProduct(ctx, &erpproductv1.UpdateProductRequest{Id: 5203, Body: &erpproductv1.UpdateProductBody{Name: "authorization probe"}})
			return response, err
		}},
		{Name: "erp products delete", App: authorization.AppERP, Permission: "erp:products:delete", Call: func(ctx context.Context) (any, error) {
			api := NewProductAPI(newAuthorizationTestProductService(ctx))
			response, err := api.DeleteProduct(ctx, &erpproductv1.DeleteProductRequest{Id: 5204})
			return response, err
		}},
	}
	testutil.RunDirectAuthorizationCases(t, cases)
}
