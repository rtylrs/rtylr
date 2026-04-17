package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/infrastructure"
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

func (s *Service) CreateEquipmentRequest(ctx context.Context, ownerID int64, params domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error) {
	if params.Quantity <= 0 {
		params.Quantity = 1
	}

	return s.repo.CreateEquipmentRequest(ctx, ownerID, params)
}

func (s *Service) ListEquipmentRequests(ctx context.Context, ownerID int64, params domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	return s.repo.ListEquipmentRequests(ctx, ownerID, params)
}

func (s *Service) ApproveEquipmentRequest(ctx context.Context, ownerID, requestID, approverID int64, approve bool, reason string) error {
	req, err := s.repo.GetEquipmentRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}

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
	return s.repo.CreateAsset(ctx, ownerID, params)
}

func (s *Service) ListAssets(ctx context.Context, ownerID int64, params domain.ListAssetsParams) ([]domain.AssetRecord, error) {
	return s.repo.ListAssets(ctx, ownerID, params)
}

func (s *Service) AssignAsset(ctx context.Context, ownerID, assetID, employeeID int64) error {
	asset, err := s.repo.GetAsset(ctx, ownerID, assetID)
	if err != nil {
		return err
	}

	if asset.Status == "assigned" {
		return domain.ErrAssetAlreadyAssigned
	}

	return s.repo.AssignAssetToEmployee(ctx, ownerID, assetID, employeeID)
}

func (s *Service) UnassignAsset(ctx context.Context, ownerID, assetID int64) error {
	_, err := s.repo.GetAsset(ctx, ownerID, assetID)
	if err != nil {
		return err
	}

	return s.repo.UnassignAsset(ctx, ownerID, assetID)
}
