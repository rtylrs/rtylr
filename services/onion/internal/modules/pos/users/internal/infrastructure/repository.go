package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos/users/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/users/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrUserNotFound is returned by FindByDisplayName when no active pos_user
// matches the (owner_id, display_name) pair.
var ErrUserNotFound = domain.ErrUserNotFound

type Repository struct {
	pool          *sql.DB
	versionReader func(context.Context, int64, int64) (int64, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

func (r *Repository) ListByOwner(ctx context.Context, businessID int64) ([]domain.POSUser, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ListPOSUsersByOwner(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.POSUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromOwnerRow(row))
	}
	return out, nil
}

func (r *Repository) ListByLocation(ctx context.Context, businessID, locationID int64) ([]domain.POSUser, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ListPOSUsersByLocation(ctx, sqlc.ListPOSUsersByLocationParams{BusinessID: businessID, LocationID: locationID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.POSUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromLocationRow(row))
	}
	return out, nil
}

// Create inserts a new pos_user. The hashedPIN argument is a bcrypt hash;
// the service layer is responsible for producing it so this layer never
// handles plaintext PINs.
func (r *Repository) Create(ctx context.Context, businessID, createdBy int64, input domain.CreatePOSUserInput, hashedPIN string) (int64, error) {
	q := sqlc.New(r.pool)
	res, err := q.CreatePOSUser(ctx, sqlc.CreatePOSUserParams{
		OwnerID:          businessID,
		BusinessID:       businessID,
		UserID:           input.UserID,
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
func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.UpdatePOSUserInput) error {
	q := sqlc.New(r.pool)
	return q.UpdatePOSUser(ctx, sqlc.UpdatePOSUserParams{
		Name:       input.Name,
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
		HourlyRate: sql.NullString{String: fmt.Sprintf("%.2f", input.HourlyRate), Valid: true},
		BusinessID: businessID,
		ID:         id,
	})
}

// UpdatePINHash stores a bcrypt hash for the given user.
func (r *Repository) UpdatePINHash(ctx context.Context, businessID, id int64, hashedPIN string) error {
	if hashedPIN == "" {
		return errors.New("refusing to write empty pin hash")
	}
	q := sqlc.New(r.pool)
	return q.UpdatePOSUserPINHash(ctx, sqlc.UpdatePOSUserPINHashParams{
		PinHash:    sql.NullString{String: hashedPIN, Valid: true},
		BusinessID: businessID,
		ID:         id,
	})
}

func (r *Repository) Deactivate(ctx context.Context, businessID, id int64) error {
	q := sqlc.New(r.pool)
	return q.DeactivatePOSUser(ctx, sqlc.DeactivatePOSUserParams{BusinessID: businessID, ID: id})
}

// FindForVerify looks up the non-secret POS-user profile used for PIN
// verification. The hash is fetched separately by GetPINHash so list/profile
// queries never carry credential material.
func (r *Repository) FindForVerify(ctx context.Context, businessID, locationID, userID int64) (*domain.POSUser, error) {
	q := sqlc.New(r.pool)
	row, err := q.FindPOSUserForPINVerify(ctx, sqlc.FindPOSUserForPINVerifyParams{BusinessID: businessID, LocationID: locationID, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u := fromVerifyRow(row)
	return &u, nil
}

func (r *Repository) GetPINHash(ctx context.Context, businessID, id int64) (string, error) {
	q := sqlc.New(r.pool)
	hash, err := q.GetPINHash(ctx, sqlc.GetPINHashParams{BusinessID: businessID, ID: id})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return hash.String, nil
}

// GetAuthorizationVersion reads the active business membership version used
// to invalidate POS credentials after access changes.
func (r *Repository) GetAuthorizationVersion(ctx context.Context, businessID, userID int64) (int64, error) {
	if r.versionReader == nil {
		return 0, errors.New("authorization version reader is not configured")
	}
	version, err := r.versionReader(ctx, businessID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}
	return version, nil
}

// FindByUserAndLocation returns a member's POS-user row at a location
// regardless of active state, used to decide whether SetPIN creates or rotates.
func (r *Repository) FindByUserAndLocation(ctx context.Context, businessID, userID, locationID int64) (*domain.POSUser, error) {
	q := sqlc.New(r.pool)
	row, err := q.FindPOSUserByUserAndLocation(ctx, sqlc.FindPOSUserByUserAndLocationParams{BusinessID: businessID, UserID: userID, LocationID: locationID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u := fromUserLocationRow(row)
	return &u, nil
}

// LocationAccess is the subset of a member's pos_member_locations row that the
// pos_users service needs to resolve role + authorisation.
type LocationAccess = domain.LocationAccess

// GetMemberLocationAccess cross-reads pos_member_locations to confirm a member
// is assigned to a location and learn their management flags. Returns
// ErrUserNotFound when the member has no access row for the location.
func (r *Repository) GetMemberLocationAccess(ctx context.Context, businessID, userID, locationID int64) (*LocationAccess, error) {
	q := sqlc.New(r.pool)
	// cross-schema read from pos_member_locations — same pos schema file
	row, err := q.GetMemberLocationAccess(ctx, sqlc.GetMemberLocationAccessParams{BusinessID: businessID, UserID: userID, LocationID: locationID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &LocationAccess{
		CanManagePOSUsers:  row.CanManagePosUsers.Bool,
		CanManageTerminals: row.CanManageTerminals.Bool,
		CanViewReports:     row.CanViewReports.Bool,
	}, nil
}

func fromOwnerRow(row sqlc.ListPOSUsersByOwnerRow) domain.POSUser {
	return domain.POSUser{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, UserID: row.UserID,
		LocationID: row.LocationID, Name: row.Name, Role: domain.POSUserRole(row.Role.PosUsersRole),
		IsActive: row.IsActive.Bool, HourlyRate: parseFloat(row.HourlyRate.String),
		MaxDiscountPct: parseFloat(row.MaxDiscountPct.String), CanVoid: row.CanVoid.Bool, CanRefund: row.CanRefund.Bool,
		CanOpenDrawer: row.CanOpenDrawer.Bool, CanApplyDiscount: row.CanApplyDiscount.Bool,
		CreatedBy: row.CreatedBy.Int64, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func fromLocationRow(row sqlc.ListPOSUsersByLocationRow) domain.POSUser {
	return domain.POSUser{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, UserID: row.UserID,
		LocationID: row.LocationID, Name: row.Name, Role: domain.POSUserRole(row.Role.PosUsersRole),
		IsActive: row.IsActive.Bool, HourlyRate: parseFloat(row.HourlyRate.String),
		MaxDiscountPct: parseFloat(row.MaxDiscountPct.String), CanVoid: row.CanVoid.Bool, CanRefund: row.CanRefund.Bool,
		CanOpenDrawer: row.CanOpenDrawer.Bool, CanApplyDiscount: row.CanApplyDiscount.Bool,
		CreatedBy: row.CreatedBy.Int64, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func fromVerifyRow(row sqlc.FindPOSUserForPINVerifyRow) domain.POSUser {
	return domain.POSUser{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, UserID: row.UserID,
		LocationID: row.LocationID, Name: row.Name, Role: domain.POSUserRole(row.Role.PosUsersRole),
		IsActive: row.IsActive.Bool, HourlyRate: parseFloat(row.HourlyRate.String),
		MaxDiscountPct: parseFloat(row.MaxDiscountPct.String), CanVoid: row.CanVoid.Bool, CanRefund: row.CanRefund.Bool,
		CanOpenDrawer: row.CanOpenDrawer.Bool, CanApplyDiscount: row.CanApplyDiscount.Bool,
		CreatedBy: row.CreatedBy.Int64, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func fromUserLocationRow(row sqlc.FindPOSUserByUserAndLocationRow) domain.POSUser {
	return domain.POSUser{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		BusinessID:       row.BusinessID,
		UserID:           row.UserID,
		LocationID:       row.LocationID,
		Name:             row.Name,
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

func (r *Repository) WithAuthorizationVersionReader(reader func(context.Context, int64, int64) (int64, error)) *Repository {
	r.versionReader = reader
	return r
}
