package booking_reminders

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_reminders/internal/api"
	"github.com/voxire/onion/internal/modules/booking_reminders/internal/application"
	bookingreminderv1 "github.com/voxire/voxie/gen/go/booking/reminder/v1"
)

type Module struct{ service *application.Service }

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := bookingreminderv1.RegisterReminderServiceServer(
		api.NewReminderAPI(m.service),
		bookingreminderv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register booking reminder service: %w", err))
	}
}
func (m *Module) GetAnonymousRoutes() []string { return []string{} }
