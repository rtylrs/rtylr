package application

import (
	"context"
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

func NewService(ctx context.Context, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(ctx),
		approvalProvider: approvalProvider,
	}
}

func invalidateHREquipmentScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(hrEquipmentScope, ownerID))
	}
}

func (s *Service) CreateEquipmentRequest(ctx context.Context, ownerID int64, params domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error) {
	if params.Quantity <= 0 {
		params.Quantity = 1
	}
	rec, err := s.repo.CreateEquipmentRequest(ctx, ownerID, params)
	if err != nil {
		return nil, err
	}
	invalidateHREquipmentScope(ctx, ownerID)
	return rec, nil
}

type listEquipmentResult struct {
	Records []domain.EquipmentRequestRecord
	Total   int64
}

func (s *Service) ListEquipmentRequests(ctx context.Context, ownerID int64, params domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	key := cacheKeys.Scoped(hrEquipmentScope, ownerID, fmt.Sprintf("reqs|e=%d|s=%s|l=%d|o=%d",
		params.EmployeeID, params.Status, params.Limit, params.Offset))
	res, err := cacher.Of[listEquipmentResult]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (listEquipmentResult, error) {
			recs, total, err := s.repo.ListEquipmentRequests(ctx, ownerID, params)
			return listEquipmentResult{Records: recs, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrEquipmentTTL),
	)
	return res.Records, res.Total, err
}

func (s *Service) ApproveEquipmentRequest(ctx context.Context, ownerID, requestID, approverID int64, approve bool, reason string) error {
	req, err := s.repo.GetEquipmentRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}
	defer invalidateHREquipmentScope(ctx, ownerID)

	if !approve {
		return s.repo.UpdateEquipmentRequestStatus(ctx, ownerID, requestID, "rejected", req.CurrentApprovalLevel, approverID, reason)
	}

	chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, ownerID, "equipment")
	if chainErr != nil {
		return chainErr
	}
	if chain == nil {
		return s.repo.UpdateEquipmentRequestStatus(ctx, ownerID, requestID, "approved", req.CurrentApprovalLevel, approverID, "")
	}
	if req.CurrentApprovalLevel+1 < chain.TotalLevels {
		return s.repo.UpdateEquipmentRequestStatus(ctx, ownerID, requestID, "pending", req.CurrentApprovalLevel+1, 0, "")
	}
	return s.repo.UpdateEquipmentRequestStatus(ctx, ownerID, requestID, "approved", req.CurrentApprovalLevel+1, approverID, "")
}

func (s *Service) CreateAsset(ctx context.Context, ownerID int64, params domain.CreateAssetParams) (*domain.AssetRecord, error) {
	rec, err := s.repo.CreateAsset(ctx, ownerID, params)
	if err != nil {
		return nil, err
	}
	invalidateHREquipmentScope(ctx, ownerID)
	return rec, nil
}

func (s *Service) ListAssets(ctx context.Context, ownerID int64, params domain.ListAssetsParams) ([]domain.AssetRecord, error) {
	loc := int64(0)
	if params.LocationID != nil {
		loc = *params.LocationID
	}
	status := ""
	if params.Status != nil {
		status = *params.Status
	}
	key := cacheKeys.Scoped(hrEquipmentScope, ownerID, fmt.Sprintf("assets|loc=%d|s=%s", loc, status))
	return cacher.Of[[]domain.AssetRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.AssetRecord, error) {
			return s.repo.ListAssets(ctx, ownerID, params)
		},
		cacher.WithExpiry(ctx, hrEquipmentTTL),
	)
}

func (s *Service) AssignAsset(ctx context.Context, ownerID, assetID, employeeID int64) error {
	asset, err := s.repo.GetAsset(ctx, ownerID, assetID)
	if err != nil {
		return err
	}
	if asset.Status == "assigned" {
		return domain.ErrAssetAlreadyAssigned
	}
	if err := s.repo.AssignAssetToEmployee(ctx, ownerID, assetID, employeeID); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, ownerID)
	return nil
}

func (s *Service) UnassignAsset(ctx context.Context, ownerID, assetID int64) error {
	if _, err := s.repo.GetAsset(ctx, ownerID, assetID); err != nil {
		return err
	}
	if err := s.repo.UnassignAsset(ctx, ownerID, assetID); err != nil {
		return err
	}
	invalidateHREquipmentScope(ctx, ownerID)
	return nil
}
