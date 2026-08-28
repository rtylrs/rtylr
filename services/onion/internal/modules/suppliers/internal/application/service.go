package application

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	"github.com/voxire/onion/internal/modules/suppliers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/audit"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type repository interface {
	ListSuppliers(ctx context.Context, businessID int64, limit, offset int) ([]domain.Supplier, bool, error)
	GetSupplierByID(ctx context.Context, businessID, id int64) (domain.Supplier, error)
	CreateSupplier(ctx context.Context, businessID int64, input domain.SupplierInput) (int64, error)
	UpdateSupplier(ctx context.Context, businessID, id int64, input domain.SupplierInput) error
	DeleteSupplier(ctx context.Context, businessID, id int64) error
}

type Service struct {
	repo repository
	db   *sql.DB
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db), db: db}
}

func paginateSuppliers(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
}

func (s *Service) ListSuppliers(ctx context.Context, businessID, page, pageSize int64) ([]domain.Supplier, bool, error) {
	limit, offset := paginateSuppliers(page, pageSize)
	items, hasMore, err := s.repo.ListSuppliers(ctx, businessID, limit, offset)
	if err != nil {
		return nil, false, err
	}
	if items == nil {
		items = []domain.Supplier{}
	}
	return items, hasMore, nil
}

func (s *Service) GetSupplierByID(ctx context.Context, businessID, id int64) (domain.Supplier, error) {
	return s.repo.GetSupplierByID(ctx, businessID, id)
}

func (s *Service) CreateSupplier(ctx context.Context, businessID int64, input domain.SupplierInput) (int64, error) {
	id, err := s.repo.CreateSupplier(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "supplier", EntityID: id, Action: "created", NewValues: input}); err != nil {
		return 0, err
	}
	s.invalidateCache(ctx, businessID)
	return id, nil
}

func (s *Service) UpdateSupplier(ctx context.Context, businessID, id int64, input domain.SupplierInput) error {
	old, err := s.repo.GetSupplierByID(ctx, businessID, id)
	if err != nil {
		return err
	}
	err = s.repo.UpdateSupplier(ctx, businessID, id, input)
	if err != nil {
		return err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "supplier", EntityID: id, Action: "updated", OldValues: old, NewValues: input}); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) DeleteSupplier(ctx context.Context, businessID, id int64) error {
	old, err := s.repo.GetSupplierByID(ctx, businessID, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteSupplier(ctx, businessID, id); err != nil {
		return err
	}
	if err := audit.Append(ctx, s.db, audit.Event{BusinessID: businessID, ActorID: businessID, EntityType: "supplier", EntityID: id, Action: "deleted", OldValues: old}); err != nil {
		return err
	}
	s.invalidateCache(ctx, businessID)
	return nil
}

func (s *Service) invalidateCache(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, fmt.Sprintf("*suppliers:%d*", businessID))
}
