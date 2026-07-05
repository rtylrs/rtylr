package security

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hubsecurityv1 "github.com/voxire/voxie/gen/go/hub/security/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/security/internal/api"
	"github.com/voxire/onion/internal/modules/security/internal/application"
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
	if err := hubsecurityv1.RegisterSecurityServiceServer(
		api.NewSecurityAPI(m.service),
		hubsecurityv1.WithMux(mux),
		hubsecurityv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register hub security service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
