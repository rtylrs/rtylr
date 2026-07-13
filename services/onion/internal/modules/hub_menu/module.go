package hubmenu

import (
	"fmt"
	"net/http"

	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/api"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubmenuv1.RegisterMenuServiceServer(
		api.NewMenuAPI(),
		hubmenuv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub menu service: %w", err))
	}
}
