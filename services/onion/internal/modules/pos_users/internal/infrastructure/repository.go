package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_users/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct{}

func NewRepository(_ context.Context) *Repository { return &Repository{} }

func (r *Repository) ListByBusiness(ctx context.Context, businessID int64) ([]domain.POSUser, error) {
	q := sqlc.New(dbclient.Pool())
	rows, err := q.ListPOSUsersByBusiness(ctx, businessID)
	if err != nil { return nil, err }
	out := make([]domain.POSUser, 0, len(rows))
	for _, row := range rows { out = append(out, fromRow(row)) }
	return out, nil
}

func (r *Repository) ListByLocation(ctx context.Context, businessID, locationID int64) ([]domain.POSUser, error) {
	q := sqlc.New(dbclient.Pool())
	rows, err := q.ListPOSUsersByLocation(ctx, sqlc.ListPOSUsersByLocationParams{BusinessID: businessID, LocationID: locationID})
	if err != nil { return nil, err }
	out := make([]domain.POSUser, 0, len(rows))
	for _, row := range rows { out = append(out, fromRow(row)) }
	return out, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, createdBy int64, input domain.CreatePOSUserInput) (int64, error) {
	q := sqlc.New(dbclient.Pool())
	res, err := q.CreatePOSUser(ctx, sqlc.CreatePOSUserParams{
		BusinessID:       businessID,
		LocationID:       input.LocationID,
		Name:             input.Name,
		Pin:              input.PIN,
		Role:             sqlc.NullPosUsersRole{PosUsersRole: sqlc.PosUsersRole(input.Role), Valid: true},
		HourlyRate:       sql.NullString{String: fmt.Sprintf("%.2f", input.HourlyRate), Valid: true},
		MaxDiscountPct:   sql.NullString{String: fmt.Sprintf("%.2f", input.MaxDiscountPct), Valid: true},
		CanVoid:          sql.NullBool{Bool: input.CanVoid, Valid: true},
		CanRefund:        sql.NullBool{Bool: input.CanRefund, Valid: true},
		CanOpenDrawer:    sql.NullBool{Bool: input.CanOpenDrawer, Valid: true},
		CanApplyDiscount: sql.NullBool{Bool: input.CanApplyDiscount, Valid: true},
		CreatedBy:        sql.NullInt64{Int64: createdBy, Valid: true},
	})
	if err != nil { return 0, err }
	id, _ := res.LastInsertId()
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.UpdatePOSUserInput) error {
	q := sqlc.New(dbclient.Pool())
	return q.UpdatePOSUser(ctx, sqlc.UpdatePOSUserParams{
		Name:             input.Name,
		Pin:              input.PIN,
		Role:             sqlc.NullPosUsersRole{PosUsersRole: sqlc.PosUsersRole(input.Role), Valid: true},
		IsActive:         sql.NullBool{Bool: input.IsActive, Valid: true},
		HourlyRate:       sql.NullString{String: fmt.Sprintf("%.2f", input.HourlyRate), Valid: true},
		MaxDiscountPct:   sql.NullString{String: fmt.Sprintf("%.2f", input.MaxDiscountPct), Valid: true},
		CanVoid:          sql.NullBool{Bool: input.CanVoid, Valid: true},
		CanRefund:        sql.NullBool{Bool: input.CanRefund, Valid: true},
		CanOpenDrawer:    sql.NullBool{Bool: input.CanOpenDrawer, Valid: true},
		CanApplyDiscount: sql.NullBool{Bool: input.CanApplyDiscount, Valid: true},
		BusinessID:       businessID,
		ID:               id,
	})
}

func (r *Repository) Deactivate(ctx context.Context, businessID, id int64) error {
	q := sqlc.New(dbclient.Pool())
	return q.DeactivatePOSUser(ctx, sqlc.DeactivatePOSUserParams{BusinessID: businessID, ID: id})
}

func (r *Repository) VerifyPIN(ctx context.Context, businessID int64, pin string) (*domain.POSUser, error) {
	q := sqlc.New(dbclient.Pool())
	row, err := q.VerifyPIN(ctx, sqlc.VerifyPINParams{BusinessID: businessID, Pin: pin})
	if err != nil {
		if err == sql.ErrNoRows { return nil, nil }
		return nil, err
	}
	u := fromRow(row)
	return &u, nil
}

func fromRow(row sqlc.PosUser) domain.POSUser {
	return domain.POSUser{
		ID:               row.ID,
		BusinessID:       row.BusinessID,
		LocationID:       row.LocationID,
		Name:             row.Name,
		PIN:              row.Pin,
		Role:             domain.POSUserRole(row.Role.PosUsersRole),
		IsActive:         row.IsActive.Bool,
		HourlyRate:       parseFloat(row.HourlyRate.String),
		MaxDiscountPct:   parseFloat(row.MaxDiscountPct.String),
		CanVoid:          row.CanVoid.Bool,
		CanRefund:        row.CanRefund.Bool,
		CanOpenDrawer:    row.CanOpenDrawer.Bool,
		CanApplyDiscount: row.CanApplyDiscount.Bool,
		CreatedBy:        row.CreatedBy.Int64,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
