package campaigns

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/campaigns/internal/api"
	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	internaldomain "github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	crmcampaignv1 "github.com/voxire/voxie/gen/go/crm/campaign/v1"
)

var ErrNotFound = errors.New("campaign not found")

// Campaign is the public persistence contract for campaign compatibility
// consumers. SQLC ownership remains within this module.
type Campaign struct {
	ID, OwnerID, BusinessID                              int64
	Name, Description, Channel, Status, Subject, Content string
	SegmentID                                            *int64
	ScheduledAt, SentAt                                  *time.Time
	AudienceCount, SentCount, DeliveredCount             int
	OpenedCount, ClickedCount                            int
	CreatedAt, UpdatedAt                                 time.Time
}

type CampaignInput struct {
	OwnerID, BusinessID                          int64
	Name, Description, Channel, Subject, Content string
	SegmentID                                    *int64
	ScheduledAt                                  *time.Time
}

type CampaignUpdateInput struct {
	ID, OwnerID, BusinessID                      int64
	Name, Description, Channel, Subject, Content string
	SegmentID                                    *int64
	ScheduledAt                                  *time.Time
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	api.RegisterRoutes(ctx, router, m.service)
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := crmcampaignv1.RegisterCampaignServiceServer(
		api.NewCampaignAPI(m.service),
		crmcampaignv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register crm campaign service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) List(ctx context.Context, businessID int64, status string) ([]Campaign, error) {
	items, err := m.service.ListByStatus(ctx, businessID, status)
	if err != nil {
		return nil, err
	}

	result := make([]Campaign, 0, len(items))
	for _, item := range items {
		result = append(result, campaignFromInternal(item))
	}

	return result, nil
}

func (m *Module) Get(ctx context.Context, businessID, id int64) (Campaign, error) {
	item, err := m.service.GetByID(ctx, businessID, id)
	if errors.Is(err, internaldomain.ErrCampaignNotFound) {
		return Campaign{}, ErrNotFound
	}

	return campaignFromInternal(item), err
}

func (m *Module) Create(ctx context.Context, input CampaignInput) (int64, error) {
	return m.service.Create(ctx, internaldomain.CampaignInput{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description,
		Type: input.Channel, AudienceSegmentID: input.SegmentID, Subject: input.Subject, Body: input.Content, ScheduledAt: input.ScheduledAt,
	})
}

func (m *Module) Update(ctx context.Context, input CampaignUpdateInput) error {
	err := m.service.Update(ctx, internaldomain.CampaignUpdateInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description,
		Type: input.Channel, AudienceSegmentID: input.SegmentID, Subject: input.Subject, Body: input.Content, ScheduledAt: input.ScheduledAt,
	})
	if errors.Is(err, internaldomain.ErrCampaignNotFound) {
		return ErrNotFound
	}

	return err
}

func (m *Module) UpdateStatus(ctx context.Context, businessID, id int64, status string) error {
	err := m.service.UpdateStatus(ctx, businessID, id, status)
	if errors.Is(err, internaldomain.ErrCampaignNotFound) {
		return ErrNotFound
	}

	return err
}

func campaignFromInternal(item internaldomain.Campaign) Campaign {
	return Campaign{
		ID: item.ID, OwnerID: item.OwnerID, BusinessID: item.BusinessID, Name: item.Name, Description: item.Description,
		Channel: item.Type, Status: item.Status, SegmentID: item.AudienceSegmentID, Subject: item.Subject, Content: item.Body,
		ScheduledAt: item.ScheduledAt, SentAt: item.SentAt, AudienceCount: item.AudienceCount, SentCount: item.SentCount,
		DeliveredCount: item.DeliveredCount, OpenedCount: item.OpenCount, ClickedCount: item.ClickedCount,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}
