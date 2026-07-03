package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	posmanagerv1 "github.com/voxire/voxie/gen/go/pos/manager/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/application"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, svc *application.Service) {
	mux := http.NewServeMux()
	if err := posmanagerv1.RegisterManagerServiceServer(
		handlers.NewManagerAPI(svc),
		posmanagerv1.WithMux(mux),
		posmanagerv1.WithErrorHandler(serverHttp.OneKitErrorHandler),
		posmanagerv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS manager service: %w", err))
	}

	router.Method(http.MethodPost, "/api/pos/manager/v1/verify", mux)
}
