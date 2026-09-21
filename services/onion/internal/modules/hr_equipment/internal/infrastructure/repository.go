package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_equipment/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_equipment/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) CreateEquipmentRequest(ctx context.Context, businessID int64, params domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error) {
	result, err := r.queries.CreateEquipmentRequest(ctx, sqlc.CreateEquipmentRequestParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		EmployeeID:   params.EmployeeID,
		ItemName:     params.ItemName,
		Quantity:     params.Quantity,
		Reason:       sqlutil.NullString(params.Reason),
		NeededByDate: nullTimeFromDateStr(params.NeededByDate),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetEquipmentRequest(ctx, businessID, id)
}

func (r *Repository) GetEquipmentRequest(ctx context.Context, businessID, id int64) (*domain.EquipmentRequestRecord, error) {
	row, err := r.queries.GetEquipmentRequest(ctx, sqlc.GetEquipmentRequestParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrEquipmentRequestNotFound
		}
		return nil, err
	}

	record := equipmentRequestFromRow(row)
	return &record, nil
}

func (r *Repository) ListEquipmentRequests(ctx context.Context, businessID int64, params domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error) {
	var total int64
	var rows []sqlc.HrEquipmentRequest
	var err error

	hasEmployee := params.EmployeeID != 0
	hasStatus := params.Status != ""

	switch {
	case hasEmployee && hasStatus:
		total, err = r.queries.CountEquipmentRequestsByEmployeeAndStatus(ctx, sqlc.CountEquipmentRequestsByEmployeeAndStatusParams{
			BusinessID: businessID, EmployeeID: params.EmployeeID, Status: params.Status,
		})
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequestsByEmployeeAndStatus(ctx, sqlc.ListEquipmentRequestsByEmployeeAndStatusParams{
			BusinessID: businessID, EmployeeID: params.EmployeeID, Status: params.Status, Limit: params.Limit, Offset: params.Offset,
		})
	case hasEmployee:
		total, err = r.queries.CountEquipmentRequestsByEmployee(ctx, sqlc.CountEquipmentRequestsByEmployeeParams{
			BusinessID: businessID, EmployeeID: params.EmployeeID,
		})
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequestsByEmployee(ctx, sqlc.ListEquipmentRequestsByEmployeeParams{
			BusinessID: businessID, EmployeeID: params.EmployeeID, Limit: params.Limit, Offset: params.Offset,
		})
	case hasStatus:
		total, err = r.queries.CountEquipmentRequestsByStatus(ctx, sqlc.CountEquipmentRequestsByStatusParams{
			BusinessID: businessID, Status: params.Status,
		})
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequestsByStatus(ctx, sqlc.ListEquipmentRequestsByStatusParams{
			BusinessID: businessID, Status: params.Status, Limit: params.Limit, Offset: params.Offset,
		})
	default:
		total, err = r.queries.CountEquipmentRequests(ctx, businessID)
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequests(ctx, sqlc.ListEquipmentRequestsParams{
			BusinessID: businessID, Limit: params.Limit, Offset: params.Offset,
		})
	}
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.EquipmentRequestRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, equipmentRequestFromRow(row))
	}

	return records, total, nil
}

func (r *Repository) UpdateEquipmentRequestStatus(ctx context.Context, businessID, id int64, status string, level int32, approvedBy int64, reason string) error {
	return r.queries.UpdateEquipmentRequestStatus(ctx, sqlc.UpdateEquipmentRequestStatusParams{
		Status:               status,
		CurrentApprovalLevel: level,
		ApprovedBy:           sqlutil.NullInt64(approvedBy),
		RejectionReason:      sqlutil.NullString(reason),
		BusinessID:           businessID,
		ID:                   id,
	})
}

func (r *Repository) CreateAsset(ctx context.Context, businessID int64, params domain.CreateAssetParams) (*domain.AssetRecord, error) {
	result, err := r.queries.CreateAsset(ctx, sqlc.CreateAssetParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		Name:          params.Name,
		Category:      params.Category,
		SerialNumber:  sqlutil.NullString(params.SerialNumber),
		LocationID:    sqlutil.NullInt64(params.LocationID),
		Status:        "available",
		PurchaseDate:  nullTimeFromDateStr(params.PurchaseDate),
		PurchasePrice: nullStringFromMinorUnits(params.PurchasePriceMinorUnits),
		Notes:         sqlutil.NullString(params.Notes),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetAsset(ctx, businessID, id)
}

func (r *Repository) GetAsset(ctx context.Context, businessID, id int64) (*domain.AssetRecord, error) {
	row, err := r.queries.GetAsset(ctx, sqlc.GetAssetParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAssetNotFound
		}
		return nil, err
	}

	record := assetFromRow(row)
	return &record, nil
}

func (r *Repository) ListAssets(ctx context.Context, businessID int64, params domain.ListAssetsParams) ([]domain.AssetRecord, int64, error) {
	filters := sqlc.ListAssetsFilteredParams{
		BusinessID: businessID,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}
	if params.LocationID != nil && *params.LocationID != 0 {
		filters.LocationID = sql.NullInt64{Int64: *params.LocationID, Valid: true}
	}
	if params.Status != nil && *params.Status != "" {
		filters.Status = sql.NullString{String: *params.Status, Valid: true}
	}
	if params.AssignedTo != nil && *params.AssignedTo != 0 {
		filters.AssignedTo = sql.NullInt64{Int64: *params.AssignedTo, Valid: true}
	}

	total, err := r.queries.CountAssetsFiltered(ctx, sqlc.CountAssetsFilteredParams{
		BusinessID: filters.BusinessID,
		LocationID: filters.LocationID,
		Status:     filters.Status,
		AssignedTo: filters.AssignedTo,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListAssetsFiltered(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	records := make([]domain.AssetRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, assetFromRow(row))
	}

	return records, total, nil
}

func (r *Repository) AssignAssetToEmployee(ctx context.Context, businessID, assetID, employeeID int64) error {
	return r.queries.AssignAsset(ctx, sqlc.AssignAssetParams{
		AssignedTo: sql.NullInt64{Int64: employeeID, Valid: true},
		BusinessID: businessID,
		ID:         assetID,
	})
}

func (r *Repository) UnassignAsset(ctx context.Context, businessID, assetID int64) error {
	return r.queries.UnassignAsset(ctx, sqlc.UnassignAssetParams{
		BusinessID: businessID,
		ID:         assetID,
	})
}

func equipmentRequestFromRow(row sqlc.HrEquipmentRequest) domain.EquipmentRequestRecord {
	var neededByDate string
	if row.NeededByDate.Valid {
		neededByDate = row.NeededByDate.Time.Format("2006-01-02")
	}

	return domain.EquipmentRequestRecord{
		ID:                   row.ID,
		OwnerID:              row.OwnerID,
		BusinessID:           row.BusinessID,
		EmployeeID:           row.EmployeeID,
		ItemName:             row.ItemName,
		Quantity:             row.Quantity,
		Reason:               row.Reason.String,
		Status:               row.Status,
		CurrentApprovalLevel: row.CurrentApprovalLevel,
		ApprovedBy:           row.ApprovedBy.Int64,
		RejectionReason:      row.RejectionReason.String,
		NeededByDate:         neededByDate,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func assetFromRow(row sqlc.HrAsset) domain.AssetRecord {
	var purchaseDate string
	if row.PurchaseDate.Valid {
		purchaseDate = row.PurchaseDate.Time.Format("2006-01-02")
	}

	return domain.AssetRecord{
		ID:                      row.ID,
		OwnerID:                 row.OwnerID,
		BusinessID:              row.BusinessID,
		Name:                    row.Name,
		Category:                row.Category,
		SerialNumber:            row.SerialNumber.String,
		LocationID:              row.LocationID.Int64,
		AssignedTo:              row.AssignedTo.Int64,
		Status:                  row.Status,
		PurchaseDate:            purchaseDate,
		PurchasePriceMinorUnits: decimalStringToMinorUnits(row.PurchasePrice.String),
		Notes:                   row.Notes.String,
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}

func nullTimeFromDateStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// nullStringFromMinorUnits converts an integer minor-units (cents) amount
// into the DECIMAL(12,2) string representation stored in hr_assets.purchase_price.
// A zero amount is stored as NULL, matching the prior float-based behavior.
func nullStringFromMinorUnits(minorUnits int64) sql.NullString {
	if minorUnits == 0 {
		return sql.NullString{Valid: false}
	}
	major := float64(minorUnits) / 100
	return sql.NullString{String: fmt.Sprintf("%.2f", major), Valid: true}
}

// decimalStringToMinorUnits converts a DECIMAL(12,2) string value read out of
// hr_assets.purchase_price into integer minor units (cents), the same
// conversion pattern used by the insights module (decimalToMinorUnits) to
// avoid representing money as a floating-point value. Empty or unparseable
// values return 0.
func decimalStringToMinorUnits(s string) int64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return int64(f*100 + 0.5)
}
