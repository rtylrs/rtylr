package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrEquipmentScope = "hr_equipment"
	hrEquipmentTTL   = 5 * time.Minute
)

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
	}
}

func invalidateHREquipmentScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrEquipmentScope, businessID))
}

func (s *Service) CreateEquipmentRequest(ctx context.Context, businessID int64, params domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error) {
	if params.Quantity <= 0 {
		params.Quantity = 1
	}
	rec, err := s.repo.CreateEquipmentRequest(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return rec, nil
}

type listEquipmentResult struct {
	Records []domain.EquipmentRequestRecord
	Total   int64
}

func (s *Service) ListEquipmentRequests(ctx context.Context, businessID int64, params domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrEquipmentScope, businessID, fmt.Sprintf("reqs|e=%d|s=%s|l=%d|o=%d",
		params.EmployeeID, params.Status, params.Limit, params.Offset))
	res, err := cacher.Of[listEquipmentResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (listEquipmentResult, error) {
			recs, total, err := s.repo.ListEquipmentRequests(ctx, businessID, params)
			return listEquipmentResult{Records: recs, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrEquipmentTTL),
	)
	return res.Records, res.Total, err
}

func (s *Service) ApproveEquipmentRequest(ctx context.Context, businessID, requestID, approverID int64, approve bool, reason string) error {
	req, err := s.repo.GetEquipmentRequest(ctx, businessID, requestID)
	if err != nil {
		return err
	}
	defer invalidateHREquipmentScope(ctx, businessID)

	if !approve {
		return s.repo.UpdateEquipmentRequestStatus(ctx, businessID, requestID, "rejected", req.CurrentApprovalLevel, approverID, reason)
	}

	chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, businessID, "equipment")
	if chainErr != nil {
		return chainErr
	}
	if chain == nil {
		return s.repo.UpdateEquipmentRequestStatus(ctx, businessID, requestID, "approved", req.CurrentApprovalLevel, approverID, "")
	}
	if req.CurrentApprovalLevel+1 < chain.TotalLevels {
		return s.repo.UpdateEquipmentRequestStatus(ctx, businessID, requestID, "pending", req.CurrentApprovalLevel+1, 0, "")
	}
	return s.repo.UpdateEquipmentRequestStatus(ctx, businessID, requestID, "approved", req.CurrentApprovalLevel+1, approverID, "")
}

func (s *Service) CreateAsset(ctx context.Context, businessID int64, params domain.CreateAssetParams) (*domain.AssetRecord, error) {
	rec, err := s.repo.CreateAsset(ctx, businessID, params)
	if err != nil {
		return nil, err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return rec, nil
}

func (s *Service) ListAssets(ctx context.Context, businessID int64, params domain.ListAssetsParams) ([]domain.AssetRecord, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	loc := int64(0)
	if params.LocationID != nil {
		loc = *params.LocationID
	}
	status := ""
	if params.Status != nil {
		status = *params.Status
	}
	key := cacheKeys.Scoped(hrEquipmentScope, businessID, fmt.Sprintf("assets|loc=%d|s=%s|l=%d|o=%d", loc, status, params.Limit, params.Offset))
	return cacher.Of[[]domain.AssetRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.AssetRecord, error) {
			return s.repo.ListAssets(ctx, businessID, params)
		},
		cacher.WithExpiry(ctx, hrEquipmentTTL),
	)
}

func (s *Service) AssignAsset(ctx context.Context, businessID, assetID, employeeID int64) error {
	asset, err := s.repo.GetAsset(ctx, businessID, assetID)
	if err != nil {
		return err
	}
	if asset.Status == "assigned" {
		return domain.ErrAssetAlreadyAssigned
	}
	if err := s.repo.AssignAssetToEmployee(ctx, businessID, assetID, employeeID); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return nil
}

func (s *Service) UnassignAsset(ctx context.Context, businessID, assetID int64) error {
	if _, err := s.repo.GetAsset(ctx, businessID, assetID); err != nil {
		return err
	}
	if err := s.repo.UnassignAsset(ctx, businessID, assetID); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, businessID)
	return nil
}
