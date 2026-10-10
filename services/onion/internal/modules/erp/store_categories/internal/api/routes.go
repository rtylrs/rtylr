package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/erp/store_categories/internal/application"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router http.LegacyRouter, service *application.Service) {
	http.RegisterGET[ListRequest, ListResponse](ctx, router, ListRoutes(), &ListHandler{Service: service})
	http.RegisterPOSTWithValidation[CreateRequest, CreateResponse](ctx, router, CreateRoutes(), &CreateHandler{Service: service})
	http.RegisterPUTWithValidation[UpdateRequest, UpdateResponse](ctx, router, UpdateRoutes(), &UpdateHandler{Service: service})
	http.RegisterDELETEWithValidation[DeleteRequest, DeleteResponse](ctx, router, DeleteRoutes(), &DeleteHandler{Service: service})
}
