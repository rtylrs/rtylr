package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	poscashv1 "github.com/voxire/voxie/gen/go/pos/cash/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/application"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, svc *application.Service) {
	mux := http.NewServeMux()
	if err := poscashv1.RegisterCashServiceServer(
		handlers.NewCashAPI(svc),
		poscashv1.WithMux(mux),
		poscashv1.WithErrorHandler(serverHttp.OneKitErrorHandler),
		poscashv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS cash service: %w", err))
	}

	router.Method(http.MethodGet, "/api/pos/cash/v1/movements", mux)
	router.Method(http.MethodPost, "/api/pos/cash/v1/drop", mux)
	router.Method(http.MethodPost, "/api/pos/cash/v1/pickup", mux)
	router.Method(http.MethodPost, "/api/pos/cash/v1/adjust", mux)
}
