package benefits

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	hrbenefitv1 "github.com/voxire/voxie/gen/go/hr/benefit/v1"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/voxire/onion/internal/modules/benefits/internal/api"
	"github.com/voxire/onion/internal/modules/benefits/internal/application"
)

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hrbenefitv1.RegisterBenefitServiceServer(
		api.NewBenefitAPI(m.service),
		hrbenefitv1.WithMux(mux),
		hrbenefitv1.WithMarshalOptions(protojson.MarshalOptions{UseProtoNames: true}),
	); err != nil {
		panic(fmt.Errorf("register hr benefit service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}
