package expenses

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	financeexpensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/expenses/internal/api"
	"github.com/voxire/onion/internal/modules/expenses/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := financeexpensesv1.RegisterExpensesServiceServer(
		api.NewExpensesAPI(m.service),
		financeexpensesv1.WithMux(mux),
		financeexpensesv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register finance expenses service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
