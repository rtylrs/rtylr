package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	moduleinfra "github.com/voxire/onion/internal/modules/notifications/internal/infrastructure"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/notifications/internal/api"
	"github.com/voxire/onion/internal/modules/notifications/internal/application"
	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
	hubnotificationv1 "github.com/voxire/voxie/gen/go/hub/notification/v1"
)

type Module struct {
	service *application.Service
}

type NotificationInput struct {
	OwnerID    int64
	BusinessID int64
	UserID     int64
	Title      string
	Body       string
	Type       string
	Link       string
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{
		service: application.NewService(ctx, moduleinfra.NewRepository(db)),
	}
}

func (m *Module) CreateNotification(ctx context.Context, input NotificationInput) error {
	return m.service.Create(ctx, domain.Notification{
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		UserID:     input.UserID,
		Title:      input.Title,
		Body:       input.Body,
		Type:       input.Type,
		Link:       input.Link,
	})
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := hubnotificationv1.RegisterNotificationServiceServer(
		api.NewNotificationAPI(m.service),
		hubnotificationv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register hub notification service: %w", err))
	}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{"/dash/ws"}
}
