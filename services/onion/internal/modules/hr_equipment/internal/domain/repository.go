package domain

import "context"

type Repository interface {
	// Equipment Requests
	CreateEquipmentRequest(ctx context.Context, businessID int64, params CreateEquipmentRequestParams) (*EquipmentRequestRecord, error)
	GetEquipmentRequest(ctx context.Context, businessID, id int64) (*EquipmentRequestRecord, error)
	ListEquipmentRequests(ctx context.Context, businessID int64, params ListEquipmentRequestsParams) ([]EquipmentRequestRecord, int64, error)
	UpdateEquipmentRequestStatus(ctx context.Context, businessID, id int64, status string, level int32, approvedBy int64, reason string) error

	// Assets
	CreateAsset(ctx context.Context, businessID int64, params CreateAssetParams) (*AssetRecord, error)
	GetAsset(ctx context.Context, businessID, id int64) (*AssetRecord, error)
	ListAssets(ctx context.Context, businessID int64, params ListAssetsParams) ([]AssetRecord, int64, error)
	AssignAssetToEmployee(ctx context.Context, businessID, assetID, employeeID int64) error
	UnassignAsset(ctx context.Context, businessID, assetID int64) error
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}
