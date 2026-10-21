package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	pool     *sql.DB
	userName func(context.Context, int64, int64) (string, error)
	member   func(context.Context, int64, int64) (bool, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{pool: db}
}

type MyLocation = domain.MyLocation

func myLocationFromValues(locationID int64, name string, address, phone sql.NullString) MyLocation {
	return MyLocation{
		LocationID: locationID,
		Name:       name,
		Address:    address.String,
		Phone:      phone.String,
	}
}

// RosterEntry is one team member on a location's POS sign-in roster.
type RosterEntry = domain.RosterEntry

func (r *Repository) GetLocationRoster(ctx context.Context, businessID, locationID int64) ([]RosterEntry, error) {
	q := sqlc.New(r.pool)
	rows, err := q.GetLocationRoster(ctx, sqlc.GetLocationRosterParams{BusinessID: businessID, LocationID: locationID})
	if err != nil {
		return nil, err
	}
	out := make([]RosterEntry, 0, len(rows))
	for _, row := range rows {
		if r.userName == nil {
			return nil, errors.New("roster name reader is not configured")
		}
		name, err := r.userName(ctx, businessID, row.UserID)
		if err != nil {
			return nil, err
		}
		out = append(out, RosterEntry{
			UserID:    row.UserID,
			Name:      name,
			HasPIN:    row.HasPin,
			IsManager: row.CanManagePosUsers.Bool,
			POSRole:   string(row.PosRole),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

func (r *Repository) ListForUser(ctx context.Context, businessID, userID int64) ([]domain.MemberLocation, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ListMemberLocations(ctx, sqlc.ListMemberLocationsParams{BusinessID: businessID, UserID: userID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.MemberLocation, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.MemberLocation{
			ID: row.ID, BusinessID: row.BusinessID, UserID: row.UserID,
			LocationID: row.LocationID, CanManagePOSUsers: row.CanManagePosUsers.Bool,
			CanManageTerminals: row.CanManageTerminals.Bool, CanViewReports: row.CanViewReports.Bool,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return out, nil
}

func (r *Repository) IsAssignedToLocation(ctx context.Context, businessID, userID, locationID int64) (bool, error) {
	var assigned bool
	err := r.pool.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pos_member_locations
			WHERE business_id = ? AND user_id = ? AND location_id = ?
		)`,
		businessID, userID, locationID,
	).Scan(&assigned)
	return assigned, err
}

func (r *Repository) GetMyLocations(ctx context.Context, businessID, userID int64) ([]MyLocation, error) {
	q := sqlc.New(r.pool)
	rows, err := q.GetMyLocations(ctx, sqlc.GetMyLocationsParams{BusinessID: businessID, UserID: userID})
	if err != nil {
		return nil, err
	}
	out := make([]MyLocation, 0, len(rows))
	for _, row := range rows {
		out = append(out, myLocationFromValues(row.LocationID, row.Name, row.Address, row.Phone))
	}
	return out, nil
}

func (r *Repository) GetBusinessLocations(ctx context.Context, businessID int64) ([]MyLocation, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ListBusinessLocations(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]MyLocation, 0, len(rows))
	for _, row := range rows {
		out = append(out, myLocationFromValues(row.LocationID, row.Name, row.Address, row.Phone))
	}
	return out, nil
}

func (r *Repository) UserBelongsToBusiness(ctx context.Context, businessID, userID int64) (bool, error) {
	if r.member == nil {
		return false, errors.New("business membership reader is not configured")
	}
	return r.member(ctx, businessID, userID)
}

func (r *Repository) LocationBelongsToBusiness(ctx context.Context, businessID, locationID int64) (bool, error) {
	q := sqlc.New(r.pool)
	return q.LocationBelongsToBusiness(ctx, sqlc.LocationBelongsToBusinessParams{BusinessID: businessID, ID: locationID})
}

func (r *Repository) Assign(ctx context.Context, businessID, userID, locationID int64, canManagePOSUsers, canManageTerminals, canViewReports bool) (int64, error) {
	q := sqlc.New(r.pool)
	res, err := q.AssignMemberToLocation(ctx, sqlc.AssignMemberToLocationParams{
		BusinessID:         businessID,
		UserID:             userID,
		LocationID:         locationID,
		CanManagePosUsers:  sql.NullBool{Bool: canManagePOSUsers, Valid: true},
		CanManageTerminals: sql.NullBool{Bool: canManageTerminals, Valid: true},
		CanViewReports:     sql.NullBool{Bool: canViewReports, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) Remove(ctx context.Context, businessID, id int64) error {
	q := sqlc.New(r.pool)
	return q.RemoveMemberFromLocation(ctx, sqlc.RemoveMemberFromLocationParams{BusinessID: businessID, ID: id})
}

func (r *Repository) WithIdentityReaders(name func(context.Context, int64, int64) (string, error), member func(context.Context, int64, int64) (bool, error)) *Repository {
	r.userName = name
	r.member = member
	return r
}
