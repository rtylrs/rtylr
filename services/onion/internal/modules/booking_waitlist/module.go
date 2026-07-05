package booking_waitlist

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	bookingwaitlistv1 "github.com/voxire/voxie/gen/go/booking/waitlist/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/api"
	"github.com/voxire/onion/internal/modules/booking_waitlist/internal/application"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingwaitlistv1.RegisterWaitlistServiceServer(
		api.NewWaitlistAPI(m.service),
		bookingwaitlistv1.WithMux(mux),
		bookingwaitlistv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register booking waitlist service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
