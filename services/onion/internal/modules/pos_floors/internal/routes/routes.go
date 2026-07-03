package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	posfloorv1 "github.com/voxire/voxie/gen/go/pos/floor/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/application"
	"github.com/voxire/onion/internal/modules/pos_floors/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, svc *application.Service) {
	mux := http.NewServeMux()
	if err := posfloorv1.RegisterFloorServiceServer(
		handlers.NewFloorAPI(svc),
		posfloorv1.WithMux(mux),
		posfloorv1.WithErrorHandler(serverHttp.OneKitErrorHandler),
		posfloorv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS floor service: %w", err))
	}

	router.Method(http.MethodGet, "/api/pos/floor/v1/floors", mux)
	router.Method(http.MethodPost, "/api/pos/floor/v1/floors", mux)
	router.Method(http.MethodPut, "/api/pos/floor/v1/floors/{id}", mux)
	router.Method(http.MethodDelete, "/api/pos/floor/v1/floors/{id}", mux)
}
