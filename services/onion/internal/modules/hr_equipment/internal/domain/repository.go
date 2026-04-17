package domain

import "context"

type Repository interface {
	// Equipment Requests
	CreateEquipmentRequest(ctx context.Context, ownerID int64, params CreateEquipmentRequestParams) (*EquipmentRequestRecord, error)
	GetEquipmentRequest(ctx context.Context, ownerID, id int64) (*EquipmentRequestRecord, error)
	ListEquipmentRequests(ctx context.Context, ownerID int64, params ListEquipmentRequestsParams) ([]EquipmentRequestRecord, int64, error)
	UpdateEquipmentRequestStatus(ctx context.Context, ownerID, id int64, status string, level int32, approvedBy int64, reason string) error

	// Assets
	CreateAsset(ctx context.Context, ownerID int64, params CreateAssetParams) (*AssetRecord, error)
	GetAsset(ctx context.Context, ownerID, id int64) (*AssetRecord, error)
	ListAssets(ctx context.Context, ownerID int64, params ListAssetsParams) ([]AssetRecord, error)
	AssignAssetToEmployee(ctx context.Context, ownerID, assetID, employeeID int64) error
	UnassignAsset(ctx context.Context, ownerID, assetID int64) error
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}
