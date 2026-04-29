package application

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/infrastructure"
)

type repository interface {
	ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error)
	UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error
	UpdateItemStatusByItemID(ctx context.Context, ownerID, itemID int64, status string) error
}

type Service struct {
	repo repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

func (s *Service) ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error) {
	return s.repo.ListActiveOrders(ctx, ownerID)
}

// ListActiveOrdersByStation returns active KDS orders routed to the given station.
// When station is empty, returns all active orders. Matching is a case-insensitive
// substring match against the order kitchen_notes field until a dedicated station
// column is added to the schema.
func (s *Service) ListActiveOrdersByStation(ctx context.Context, ownerID int64, station string) ([]domain.KDSOrder, error) {
	orders, err := s.repo.ListActiveOrders(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if station == "" {
		return orders, nil
	}
	needle := strings.ToLower(strings.TrimSpace(station))
	filtered := make([]domain.KDSOrder, 0, len(orders))
	for _, o := range orders {
		if strings.Contains(strings.ToLower(o.KitchenNotes), needle) {
			filtered = append(filtered, o)
		}
	}
	return filtered, nil
}

func (s *Service) UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error {
	return s.repo.UpdateItemStatus(ctx, input)
}

func (s *Service) UpdateItemStatusByItemID(ctx context.Context, ownerID, itemID int64, status string) error {
	return s.repo.UpdateItemStatusByItemID(ctx, ownerID, itemID, status)
}
