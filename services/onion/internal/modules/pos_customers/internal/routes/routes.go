package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	poscustomerv1 "github.com/voxire/voxie/gen/go/pos/customer/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/application"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/routes/handlers"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(_ context.Context, router *chi.Mux, svc *application.Service) {
	mux := http.NewServeMux()
	if err := poscustomerv1.RegisterCustomerServiceServer(
		handlers.NewCustomerAPI(svc),
		poscustomerv1.WithMux(mux),
		poscustomerv1.WithErrorHandler(serverHttp.OneKitErrorHandler),
		poscustomerv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS customer service: %w", err))
	}

	router.Method(http.MethodGet, "/api/pos/customer/v1/customers", mux)
}
