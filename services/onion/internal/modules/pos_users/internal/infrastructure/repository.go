package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_users/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrUserNotFound is returned by FindByDisplayName when no active pos_user
// matches the (owner_id, display_name) pair.
var ErrUserNotFound = errors.New("pos user not found")

type Repository struct{}

func NewRepository(_ context.Context) *Repository { return &Repository{} }

func (r *Repository) ListByOwner(ctx context.Context, ownerID int64) ([]domain.POSUser, error) {
	q := sqlc.New(dbclient.Pool())
	rows, err := q.ListPOSUsersByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.POSUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

func (r *Repository) ListByLocation(ctx context.Context, ownerID, locationID int64) ([]domain.POSUser, error) {
	q := sqlc.New(dbclient.Pool())
	rows, err := q.ListPOSUsersByLocation(ctx, sqlc.ListPOSUsersByLocationParams{OwnerID: ownerID, LocationID: locationID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.POSUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}

// Create inserts a new pos_user. The hashedPIN argument is a bcrypt hash;
// the service layer is responsible for producing it so this layer never
// handles plaintext PINs.
func (r *Repository) Create(ctx context.Context, ownerID int64, createdBy int64, input domain.CreatePOSUserInput, hashedPIN string) (int64, error) {
	q := sqlc.New(dbclient.Pool())
	res, err := q.CreatePOSUser(ctx, sqlc.CreatePOSUserParams{
		OwnerID:          ownerID,
		LocationID:       input.LocationID,
		Name:             input.Name,
		PinHash:          sql.NullString{String: hashedPIN, Valid: hashedPIN != ""},
		Role:             sqlc.NullPosUsersRole{PosUsersRole: sqlc.PosUsersRole(input.Role), Valid: true},
		HourlyRate:       sql.NullString{String: fmt.Sprintf("%.2f", input.HourlyRate), Valid: true},
		MaxDiscountPct:   sql.NullString{String: fmt.Sprintf("%.2f", input.MaxDiscountPct), Valid: true},
		CanVoid:          sql.NullBool{Bool: input.CanVoid, Valid: true},
		CanRefund:        sql.NullBool{Bool: input.CanRefund, Valid: true},
		CanOpenDrawer:    sql.NullBool{Bool: input.CanOpenDrawer, Valid: true},
		CanApplyDiscount: sql.NullBool{Bool: input.CanApplyDiscount, Valid: true},
		CreatedBy:        sql.NullInt64{Int64: createdBy, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update mutates non-PIN fields on a pos_user. PIN rotation goes through
// UpdatePINHash, which enforces that only a bcrypt hash can be written.
func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.UpdatePOSUserInput) error {
	q := sqlc.New(dbclient.Pool())
	return q.UpdatePOSUser(ctx, sqlc.UpdatePOSUserParams{
		Name:             input.Name,
		Role:             sqlc.NullPosUsersRole{PosUsersRole: sqlc.PosUsersRole(input.Role), Valid: true},
		IsActive:         sql.NullBool{Bool: input.IsActive, Valid: true},
		HourlyRate:       sql.NullString{String: fmt.Sprintf("%.2f", input.HourlyRate), Valid: true},
		MaxDiscountPct:   sql.NullString{String: fmt.Sprintf("%.2f", input.MaxDiscountPct), Valid: true},
		CanVoid:          sql.NullBool{Bool: input.CanVoid, Valid: true},
		CanRefund:        sql.NullBool{Bool: input.CanRefund, Valid: true},
		CanOpenDrawer:    sql.NullBool{Bool: input.CanOpenDrawer, Valid: true},
		CanApplyDiscount: sql.NullBool{Bool: input.CanApplyDiscount, Valid: true},
		OwnerID:          ownerID,
		ID:               id,
	})
}

// UpdatePINHash stores a bcrypt hash for the given user and clears the
// legacy plaintext column in the same statement.
func (r *Repository) UpdatePINHash(ctx context.Context, ownerID, id int64, hashedPIN string) error {
	if hashedPIN == "" {
		return errors.New("refusing to write empty pin hash")
	}
	q := sqlc.New(dbclient.Pool())
	return q.UpdatePOSUserPINHash(ctx, sqlc.UpdatePOSUserPINHashParams{
		PinHash: sql.NullString{String: hashedPIN, Valid: true},
		OwnerID: ownerID,
		ID:      id,
	})
}

func (r *Repository) Deactivate(ctx context.Context, ownerID, id int64) error {
	q := sqlc.New(dbclient.Pool())
	return q.DeactivatePOSUser(ctx, sqlc.DeactivatePOSUserParams{OwnerID: ownerID, ID: id})
}

// FindByDisplayName looks up a single active pos_user by owner + display
// name. The row carries both the legacy plaintext PIN and the hashed PIN
// so the service layer can decide how to verify; the caller is expected to
// zero those fields before returning anything to the transport layer.
func (r *Repository) FindByDisplayName(ctx context.Context, ownerID int64, displayName string) (*domain.POSUser, error) {
	q := sqlc.New(dbclient.Pool())
	row, err := q.FindPOSUserForPINVerify(ctx, sqlc.FindPOSUserForPINVerifyParams{OwnerID: ownerID, Name: displayName})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u := fromRow(row)
	return &u, nil
}

func fromRow(row sqlc.PosUser) domain.POSUser {
	return domain.POSUser{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		LocationID:       row.LocationID,
		Name:             row.Name,
		PIN:              row.Pin,
		PINHash:          row.PinHash.String,
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
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from pos_users row",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return f
}
