package domain

import "time"

type EquipmentRequestRecord struct {
	ID                   int64
	OwnerID              int64
	BusinessID           int64
	EmployeeID           int64
	ItemName             string
	Quantity             int32
	Reason               string
	Status               string
	CurrentApprovalLevel int32
	ApprovedBy           int64
	RejectionReason      string
	NeededByDate         string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type AssetRecord struct {
	ID                      int64
	OwnerID                 int64
	BusinessID              int64
	Name                    string
	Category                string
	SerialNumber            string
	LocationID              int64
	AssignedTo              int64
	Status                  string
	Condition               string
	PurchaseDate            string
	PurchasePriceMinorUnits int64
	Notes                   string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type ListEquipmentRequestsParams struct {
	EmployeeID int64
	Status     string
	Limit      int32
	Offset     int32
}

type CreateEquipmentRequestParams struct {
	EmployeeID   int64
	ItemName     string
	Quantity     int32
	Reason       string
	NeededByDate string
}

type CreateAssetParams struct {
	Name                    string
	Category                string
	SerialNumber            string
	LocationID              int64
	Condition               string
	PurchaseDate            string
	PurchasePriceMinorUnits int64
	Notes                   string
}

type ListAssetsParams struct {
	LocationID *int64
	// Status is retained for the legacy lifecycle filter.
	Status *string
	// Condition is the physical-condition filter used by the generated API.
	Condition  *string
	AssignedTo *int64
	Limit      int32
	Offset     int32
}

type ApprovalChainInfo struct {
	TotalLevels int32
	Levels      []ApprovalLevelInfo
}

type ApprovalLevelInfo struct {
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
}
