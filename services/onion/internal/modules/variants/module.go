package variants

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	erpvariantv1 "github.com/voxire/voxie/gen/go/erp/variant/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/variants/internal/api"
	"github.com/voxire/onion/internal/modules/variants/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, db),
	}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := erpvariantv1.RegisterVariantServiceServer(
		api.NewVariantAPI(m.service),
		erpvariantv1.WithMux(mux),
		erpvariantv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register ERP variant service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
