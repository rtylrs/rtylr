package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, svc *application.Service) {
	mux := http.NewServeMux()
	if err := possummaryv1.RegisterDaySummaryServiceServer(
		handlers.NewDaySummaryAPI(svc),
		possummaryv1.WithMux(mux),
		possummaryv1.WithErrorHandler(serverHttp.OneKitErrorHandler),
		possummaryv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS day summary service: %w", err))
	}

	router.Method(http.MethodGet, "/api/pos/day-summary/v1/day-summary", mux)
}
