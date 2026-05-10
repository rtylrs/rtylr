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
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) CreateEquipmentRequest(ctx context.Context, ownerID int64, params domain.CreateEquipmentRequestParams) (*domain.EquipmentRequestRecord, error) {
	result, err := r.queries.CreateEquipmentRequest(ctx, sqlc.CreateEquipmentRequestParams{
		OwnerID:      ownerID,
		EmployeeID:   params.EmployeeID,
		ItemName:     params.ItemName,
		Quantity:     params.Quantity,
		Reason:       nullString(params.Reason),
		NeededByDate: nullTimeFromDateStr(params.NeededByDate),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetEquipmentRequest(ctx, ownerID, id)
}

func (r *Repository) GetEquipmentRequest(ctx context.Context, ownerID, id int64) (*domain.EquipmentRequestRecord, error) {
	row, err := r.queries.GetEquipmentRequest(ctx, sqlc.GetEquipmentRequestParams{
		OwnerID: ownerID,
		ID:      id,
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

func (r *Repository) ListEquipmentRequests(ctx context.Context, ownerID int64, params domain.ListEquipmentRequestsParams) ([]domain.EquipmentRequestRecord, int64, error) {
	var total int64
	var rows []sqlc.HrEquipmentRequest
	var err error

	hasEmployee := params.EmployeeID != 0
	hasStatus := params.Status != ""

	switch {
	case hasEmployee && hasStatus:
		total, err = r.queries.CountEquipmentRequestsByEmployeeAndStatus(ctx, sqlc.CountEquipmentRequestsByEmployeeAndStatusParams{
			OwnerID: ownerID, EmployeeID: params.EmployeeID, Status: params.Status,
		})
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequestsByEmployeeAndStatus(ctx, sqlc.ListEquipmentRequestsByEmployeeAndStatusParams{
			OwnerID: ownerID, EmployeeID: params.EmployeeID, Status: params.Status, Limit: params.Limit, Offset: params.Offset,
		})
	case hasEmployee:
		total, err = r.queries.CountEquipmentRequestsByEmployee(ctx, sqlc.CountEquipmentRequestsByEmployeeParams{
			OwnerID: ownerID, EmployeeID: params.EmployeeID,
		})
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequestsByEmployee(ctx, sqlc.ListEquipmentRequestsByEmployeeParams{
			OwnerID: ownerID, EmployeeID: params.EmployeeID, Limit: params.Limit, Offset: params.Offset,
		})
	case hasStatus:
		total, err = r.queries.CountEquipmentRequestsByStatus(ctx, sqlc.CountEquipmentRequestsByStatusParams{
			OwnerID: ownerID, Status: params.Status,
		})
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequestsByStatus(ctx, sqlc.ListEquipmentRequestsByStatusParams{
			OwnerID: ownerID, Status: params.Status, Limit: params.Limit, Offset: params.Offset,
		})
	default:
		total, err = r.queries.CountEquipmentRequests(ctx, ownerID)
		if err != nil {
			return nil, 0, err
		}
		rows, err = r.queries.ListEquipmentRequests(ctx, sqlc.ListEquipmentRequestsParams{
			OwnerID: ownerID, Limit: params.Limit, Offset: params.Offset,
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

func (r *Repository) UpdateEquipmentRequestStatus(ctx context.Context, ownerID, id int64, status string, level int32, approvedBy int64, reason string) error {
	return r.queries.UpdateEquipmentRequestStatus(ctx, sqlc.UpdateEquipmentRequestStatusParams{
		Status:               status,
		CurrentApprovalLevel: level,
		ApprovedBy:           nullInt64(approvedBy),
		RejectionReason:      nullString(reason),
		OwnerID:              ownerID,
		ID:                   id,
	})
}

func (r *Repository) CreateAsset(ctx context.Context, ownerID int64, params domain.CreateAssetParams) (*domain.AssetRecord, error) {
	result, err := r.queries.CreateAsset(ctx, sqlc.CreateAssetParams{
		OwnerID:       ownerID,
		Name:          params.Name,
		Category:      params.Category,
		SerialNumber:  nullString(params.SerialNumber),
		LocationID:    nullInt64(params.LocationID),
		Status:        "available",
		PurchaseDate:  nullTimeFromDateStr(params.PurchaseDate),
		PurchasePrice: nullStringFromFloat(params.PurchasePrice),
		Notes:         nullString(params.Notes),
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetAsset(ctx, ownerID, id)
}

func (r *Repository) GetAsset(ctx context.Context, ownerID, id int64) (*domain.AssetRecord, error) {
	row, err := r.queries.GetAsset(ctx, sqlc.GetAssetParams{
		OwnerID: ownerID,
		ID:      id,
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

func (r *Repository) ListAssets(ctx context.Context, ownerID int64, params domain.ListAssetsParams) ([]domain.AssetRecord, error) {
	var rows []sqlc.HrAsset
	var err error

	hasLocation := params.LocationID != nil && *params.LocationID != 0
	hasStatus := params.Status != nil && *params.Status != ""

	switch {
	case hasLocation && hasStatus:
		rows, err = r.queries.ListAssetsByLocationAndStatus(ctx, sqlc.ListAssetsByLocationAndStatusParams{
			OwnerID: ownerID, LocationID: sql.NullInt64{Int64: *params.LocationID, Valid: true}, Status: *params.Status,
		})
	case hasLocation:
		rows, err = r.queries.ListAssetsByLocation(ctx, sqlc.ListAssetsByLocationParams{
			OwnerID: ownerID, LocationID: sql.NullInt64{Int64: *params.LocationID, Valid: true},
		})
	case hasStatus:
		rows, err = r.queries.ListAssetsByStatus(ctx, sqlc.ListAssetsByStatusParams{
			OwnerID: ownerID, Status: *params.Status,
		})
	default:
		rows, err = r.queries.ListAssets(ctx, ownerID)
	}
	if err != nil {
		return nil, err
	}

	records := make([]domain.AssetRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, assetFromRow(row))
	}

	return records, nil
}

func (r *Repository) AssignAssetToEmployee(ctx context.Context, ownerID, assetID, employeeID int64) error {
	return r.queries.AssignAsset(ctx, sqlc.AssignAssetParams{
		AssignedTo: sql.NullInt64{Int64: employeeID, Valid: true},
		OwnerID:    ownerID,
		ID:         assetID,
	})
}

func (r *Repository) UnassignAsset(ctx context.Context, ownerID, assetID int64) error {
	return r.queries.UnassignAsset(ctx, sqlc.UnassignAssetParams{
		OwnerID: ownerID,
		ID:      assetID,
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
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		Name:          row.Name,
		Category:      row.Category,
		SerialNumber:  row.SerialNumber.String,
		LocationID:    row.LocationID.Int64,
		AssignedTo:    row.AssignedTo.Int64,
		Status:        row.Status,
		PurchaseDate:  purchaseDate,
		PurchasePrice: parseFloat(row.PurchasePrice.String),
		Notes:         row.Notes.String,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
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

func nullStringFromFloat(f float64) sql.NullString {
	if f == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: fmt.Sprintf("%.2f", f), Valid: true}
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return f
}
