package hr_performance

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/api"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterGET[api.GetConfigRequest, api.GetConfigResponse](
		ctx, router, api.GetConfigRoutes(), &api.GetConfigHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpsertConfigRequest, api.UpsertConfigResponse](
		ctx, router, api.UpsertConfigRoutes(), &api.UpsertConfigHandler{Service: m.service},
	)
	serverHttp.RegisterGET[api.GetCycleRequest, api.GetCycleResponse](
		ctx, router, api.GetCycleRoutes(), &api.GetCycleHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.OpenCycleRequest, api.OpenCycleResponse](
		ctx, router, api.OpenCycleRoutes(), &api.OpenCycleHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.CreateGoalRequest, api.CreateGoalResponse](
		ctx, router, api.CreateGoalRoutes(), &api.CreateGoalHandler{Service: m.service},
	)
	serverHttp.RegisterPUTWithValidation[api.UpdateGoalRequest, api.UpdateGoalResponse](
		ctx, router, api.UpdateGoalRoutes(), &api.UpdateGoalHandler{Service: m.service},
	)
	serverHttp.RegisterGET[api.ListReviewsRequest, api.ListReviewsResponse](
		ctx, router, api.ListReviewsRoutes(), &api.ListReviewsHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.SubmitReviewRequest, api.SubmitReviewResponse](
		ctx, router, api.SubmitReviewRoutes(), &api.SubmitReviewHandler{Service: m.service},
	)
	serverHttp.RegisterPOSTWithValidation[api.ShareReviewRequest, api.ShareReviewResponse](
		ctx, router, api.ShareReviewRoutes(), &api.ShareReviewHandler{Service: m.service},
	)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	err := hrperformancev1.RegisterPerformanceServiceServer(
		api.NewPerformanceService(m.service),
		hrperformancev1.WithMux(mux),
		hrperformancev1.WithMiddleware(api.PerformancePaginationMiddleware),
	)
	if err != nil {
		panic(err)
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
