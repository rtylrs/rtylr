package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) LinkInviteAcceptance(ctx context.Context, businessID, employeeID, userID int64) error {
	result, err := r.queries.LinkInviteAcceptance(ctx, sqlc.LinkInviteAcceptanceParams{
		UserID:     sqlutil.NullInt64(userID),
		ID:         employeeID,
		BusinessID: businessID,
	})
	if err != nil {
		if isDuplicateUserIDError(err) {
			return domain.ErrUserAlreadyLinked
		}
		logger.Ctx(ctx).Error("failed to link invite acceptance", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return domain.ErrEmployeeNotFound
	}
	return nil
}

func (r *Repository) CreateEmployeeWithBalances(ctx context.Context, businessID int64, emp domain.EmployeeRecord, leaveTypeIDs []int64) (*domain.EmployeeRecord, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	result, err := qtx.CreateEmployee(ctx, employeeToParams(businessID, emp))
	if err != nil {
		_ = tx.Rollback()
		if isDuplicateUserIDError(err) {
			return nil, domain.ErrUserAlreadyLinked
		}
		logger.Ctx(ctx).Error("failed to create employee in tx", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}

	employeeID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	currentYear := int32(time.Now().Year()) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	for _, leaveTypeID := range leaveTypeIDs {
		err := qtx.CreateLeaveBalance(ctx, sqlc.CreateLeaveBalanceParams{
			OwnerID:     businessID,
			BusinessID:  businessID,
			EmployeeID:  employeeID,
			LeaveTypeID: leaveTypeID,
			Year:        currentYear,
		})
		if err != nil {
			_ = tx.Rollback()
			logger.Ctx(ctx).Error("failed to create leave balance", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	emp.ID = employeeID
	emp.OwnerID = businessID
	emp.BusinessID = businessID
	return &emp, nil
}

// --- Invites ---

func (r *Repository) CreateInviteToken(ctx context.Context, businessID, employeeID int64, expiresAt time.Time) (string, error) {
	token := uuid.New().String()
	_, err := r.queries.CreateInviteToken(ctx, sqlc.CreateInviteTokenParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		EmployeeID: employeeID,
		Token:      token,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to create invite token", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return "", err
	}
	return token, nil
}

func (r *Repository) GetInviteByToken(ctx context.Context, token string) (*domain.InviteRecord, error) {
	row, err := r.queries.GetInviteByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInviteTokenInvalid
		}
		return nil, err
	}
	return &domain.InviteRecord{
		ID:         row.ID,
		OwnerID:    row.OwnerID,
		BusinessID: row.BusinessID,
		EmployeeID: row.EmployeeID,
		Token:      row.Token,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
	}, nil
}

func (r *Repository) GetAccountEmail(ctx context.Context, userID int64) (string, error) {
	if r.accountEmail == nil {
		return "", errors.New("account email reader is not configured")
	}
	email, err := r.accountEmail(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrInviteRecipientMismatch
		}
		logger.Ctx(ctx).Error("failed to read account email for invite acceptance", logger.Any("user_id", userID), logger.Error(err))
		return "", err
	}
	return email, nil
}

// AcceptInvite consumes the invite token and links the user to the invited
// employee in a single transaction. Consuming the token before the link would
// burn it when the link cannot succeed - the user_id already belongs to
// another employee row, or the employee row is gone - leaving the invitee with
// "invalid invite token" on every retry and the employee still unlinked.
func (r *Repository) AcceptInvite(ctx context.Context, businessID, employeeID, userID int64, token string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	consumed, err := qtx.ConsumeInviteToken(ctx, token)
	if err != nil {
		_ = tx.Rollback()
		logger.Ctx(ctx).Error("failed to consume invite token", logger.Any("business_id", businessID), logger.Error(err))
		return err
	}
	if affected, err := consumed.RowsAffected(); err != nil {
		_ = tx.Rollback()
		return err
	} else if affected == 0 {
		_ = tx.Rollback()
		return domain.ErrInviteTokenInvalid
	}

	linked, err := qtx.LinkInviteAcceptance(ctx, sqlc.LinkInviteAcceptanceParams{
		UserID:     sqlutil.NullInt64(userID),
		ID:         employeeID,
		BusinessID: businessID,
	})
	if err != nil {
		_ = tx.Rollback()
		if isDuplicateUserIDError(err) {
			return domain.ErrUserAlreadyLinked
		}
		logger.Ctx(ctx).Error("failed to link invite acceptance", logger.Any("business_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return err
	}
	if affected, err := linked.RowsAffected(); err != nil {
		_ = tx.Rollback()
		return err
	} else if affected == 0 {
		_ = tx.Rollback()
		return domain.ErrEmployeeNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// --- Locations ---

func (r *Repository) ListLocations(ctx context.Context, businessID int64) ([]domain.LocationRecord, error) {
	rows, err := r.queries.ListLocations(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("failed to list locations", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	locations := make([]domain.LocationRecord, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, locationFromRow(row))
	}
	return locations, nil
}

func (r *Repository) GetLocation(ctx context.Context, businessID, locationID int64) (*domain.LocationRecord, error) {
	row, err := r.queries.GetLocation(ctx, sqlc.GetLocationParams{BusinessID: businessID, ID: locationID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}
	loc := locationFromRow(row)
	return &loc, nil
}
