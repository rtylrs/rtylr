package segments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/segments/internal/api"
	"github.com/voxire/onion/internal/modules/segments/internal/application"
	internaldomain "github.com/voxire/onion/internal/modules/segments/internal/domain"
	crmsegmentv1 "github.com/voxire/voxie/gen/go/crm/segment/v1"
)

var ErrNotFound = errors.New("segment not found")

// Segment is the public persistence contract for segment compatibility
// consumers. SQLC ownership remains within this module.
type Segment struct {
	ID, OwnerID, BusinessID int64
	Name, Description       string
	Rules                   []byte
	CustomerCount           int
	LastCalculatedAt        *time.Time
	CreatedAt, UpdatedAt    time.Time
}

type SegmentInput struct {
	OwnerID, BusinessID int64
	Name, Description   string
	Rules               []byte
}

type SegmentUpdateInput struct {
	ID, OwnerID, BusinessID int64
	Name, Description       string
	Rules                   []byte
}

type Module struct {
	service *application.Service
}

func NewModule(ctx context.Context, db *sql.DB) *Module {
	return &Module{service: application.NewService(ctx, db)}
}

func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := crmsegmentv1.RegisterSegmentServiceServer(
		api.NewSegmentAPI(m.service),
		crmsegmentv1.WithMux(mux),
	); err != nil {
		panic(fmt.Errorf("register segment service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func (m *Module) List(ctx context.Context, businessID int64) ([]Segment, error) {
	items, err := m.service.List(ctx, businessID)
	if err != nil {
		return nil, err
	}

	result := make([]Segment, 0, len(items))
	for _, item := range items {
		result = append(result, segmentFromInternal(item))
	}

	return result, nil
}

func (m *Module) Get(ctx context.Context, businessID, id int64) (Segment, error) {
	item, err := m.service.Get(ctx, businessID, id)
	if errors.Is(err, internaldomain.ErrSegmentNotFound) {
		return Segment{}, ErrNotFound
	}

	return segmentFromInternal(item), err
}

func (m *Module) Create(ctx context.Context, input SegmentInput) (int64, error) {
	return m.service.Create(ctx, internaldomain.SegmentInput{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Filters: input.Rules,
	})
}

func (m *Module) Update(ctx context.Context, input SegmentUpdateInput) error {
	err := m.service.Update(ctx, internaldomain.SegmentUpdateInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Filters: input.Rules,
	})
	if errors.Is(err, internaldomain.ErrSegmentNotFound) {
		return ErrNotFound
	}

	return err
}

func (m *Module) Delete(ctx context.Context, businessID, id int64) error {
	if _, err := m.Get(ctx, businessID, id); err != nil {
		return err
	}

	return m.service.Delete(ctx, businessID, id)
}

func segmentFromInternal(item internaldomain.Segment) Segment {
	return Segment{
		ID: item.ID, OwnerID: item.OwnerID, BusinessID: item.BusinessID, Name: item.Name, Description: item.Description,
		Rules: item.Filters, CustomerCount: item.CustomerCount, LastCalculatedAt: item.LastCalculatedAt,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}
