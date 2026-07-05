package financereports

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financestatementv1 "github.com/voxire/voxie/gen/go/finance/statement/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/api"
	"github.com/voxire/onion/internal/modules/finance_reports/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financestatementv1.RegisterStatementServiceServer(
		api.NewStatementAPI(m.service),
		financestatementv1.WithMux(mux),
		financestatementv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register finance statement service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
