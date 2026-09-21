package domain

import "time"

type EquipmentRequestRecord struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	EmployeeID           int64     `json:"employee_id"`
	ItemName             string    `json:"item_name"`
	Quantity             int32     `json:"quantity"`
	Reason               string    `json:"reason"`
	Status               string    `json:"status"`
	CurrentApprovalLevel int32     `json:"current_approval_level"`
	ApprovedBy           int64     `json:"approved_by"`
	RejectionReason      string    `json:"rejection_reason"`
	NeededByDate         string    `json:"needed_by_date"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type AssetRecord struct {
	ID                      int64     `json:"id"`
	OwnerID                 int64     `json:"owner_id"`
	BusinessID              int64     `json:"business_id"`
	Name                    string    `json:"name"`
	Category                string    `json:"category"`
	SerialNumber            string    `json:"serial_number"`
	LocationID              int64     `json:"location_id"`
	AssignedTo              int64     `json:"assigned_to"`
	Status                  string    `json:"status"`
	PurchaseDate            string    `json:"purchase_date"`
	PurchasePriceMinorUnits int64     `json:"purchase_price_minor_units"`
	Notes                   string    `json:"notes"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type ListEquipmentRequestsParams struct {
	EmployeeID int64
	Status     string
	Limit      int32
	Offset     int32
}

type CreateEquipmentRequestParams struct {
	EmployeeID   int64  `json:"employee_id"`
	ItemName     string `json:"item_name"`
	Quantity     int32  `json:"quantity"`
	Reason       string `json:"reason"`
	NeededByDate string `json:"needed_by_date"`
}

type CreateAssetParams struct {
	Name                    string `json:"name"`
	Category                string `json:"category"`
	SerialNumber            string `json:"serial_number"`
	LocationID              int64  `json:"location_id"`
	PurchaseDate            string `json:"purchase_date"`
	PurchasePriceMinorUnits int64  `json:"purchase_price_minor_units"`
	Notes                   string `json:"notes"`
}

type ListAssetsParams struct {
	LocationID *int64
	Status     *string
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
