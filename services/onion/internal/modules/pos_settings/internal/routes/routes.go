package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	possettingsv1 "github.com/voxire/voxie/gen/go/pos/settings/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_settings/internal/application"
	"github.com/voxire/onion/internal/modules/pos_settings/internal/routes/handlers"
	contractHttp "github.com/voxire/onion/internal/platform/contracts"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, svc *application.Service) {
	mux := http.NewServeMux()
	if err := possettingsv1.RegisterSettingsServiceServer(
		handlers.NewSettingsAPI(svc),
		possettingsv1.WithMux(mux),
		possettingsv1.WithErrorHandler(contractHttp.ErrorHandler),
		possettingsv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS settings service: %w", err))
	}

	router.Method(http.MethodGet, "/api/pos/settings/v1/settings", mux)
	router.Method(http.MethodPut, "/api/pos/settings/v1/settings", mux)
}
