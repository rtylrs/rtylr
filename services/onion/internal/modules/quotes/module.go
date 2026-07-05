package quotes

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/quotes/internal/api"
	"github.com/voxire/onion/internal/modules/quotes/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	financequotev1 "github.com/voxire/voxie/gen/go/finance/quote/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[api.ListRequest, api.ListResponse](
		ctx, router, api.ListRoutes(), &api.ListHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CreateRequest, api.CreateResponse](
		ctx, router, api.CreateRoutes(), &api.CreateHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateRequest, api.UpdateResponse](
		ctx, router, api.UpdateRoutes(), &api.UpdateHandler{Service: m.service},
	)
	serverHttp.RegisterDELETEWithValidation[api.DeleteRequest, api.DeleteResponse](
		ctx, router, api.DeleteRoutes(), &api.DeleteHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateQuoteByIDRequest, api.UpdateQuoteByIDResponse](
		ctx, router, api.UpdateQuoteByIDRoutes(), &api.UpdateQuoteByIDHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateQuoteStatusRequest, api.UpdateQuoteStatusResponse](
		ctx, router, api.UpdateQuoteStatusRoutes(), &api.UpdateQuoteStatusHandler{Service: m.service},
	)
	serverHttp.RegisterDELETEWithValidation[api.DeleteQuoteByIDRequest, api.DeleteQuoteByIDResponse](
		ctx, router, api.DeleteQuoteByIDRoutes(), &api.DeleteQuoteByIDHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.ConvertQuoteRequest, api.ConvertQuoteResponse](
		ctx, router, api.ConvertQuoteRoutes(), &api.ConvertQuoteHandler{Service: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financequotev1.RegisterQuoteServiceServer(
		api.NewQuoteAPI(m.service),
		financequotev1.WithMux(mux),
		financequotev1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register finance quote service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
