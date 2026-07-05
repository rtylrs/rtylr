package pos_day_summary

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	possummaryv1 "github.com/voxire/voxie/gen/go/pos/day_summary/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/api"
	"github.com/voxire/onion/internal/modules/pos_day_summary/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := possummaryv1.RegisterDaySummaryServiceServer(
		api.NewDaySummaryAPI(m.service),
		possummaryv1.WithMux(mux),
		possummaryv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register POS day summary service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
