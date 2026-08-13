package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/custom_domains/internal/domain"
	"github.com/voxire/onion/internal/modules/custom_domains/internal/infrastructure/persistence/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(db *sql.DB) *Repository { return &Repository{queries: sqlc.New(db)} }

func (r *Repository) Get(ctx context.Context, businessID int64) (domain.CustomDomain, error) {
	row, err := r.queries.GetCustomDomain(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CustomDomain{}, domain.ErrNotFound
	}
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return domain.CustomDomain{}, domain.ErrConflict
		}
		return domain.CustomDomain{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) Register(ctx context.Context, input domain.CustomDomain) (domain.CustomDomain, error) {
	result, err := r.queries.RegisterCustomDomain(ctx, sqlc.RegisterCustomDomainParams{OwnerID: input.OwnerID, BusinessID: input.BusinessID, Domain: input.Domain, VerificationToken: input.VerificationToken})
	if err != nil {
		return domain.CustomDomain{}, err
	}
	if _, err := result.LastInsertId(); err != nil {
		return domain.CustomDomain{}, err
	}
	return r.Get(ctx, input.BusinessID)
}

func (r *Repository) Verify(ctx context.Context, businessID int64, token string) error {
	if err := r.queries.VerifyCustomDomain(ctx, sqlc.VerifyCustomDomainParams{BusinessID: businessID, VerificationToken: token}); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Remove(ctx context.Context, businessID int64) error {
	return r.queries.RemoveCustomDomain(ctx, businessID)
}

func fromRow(row sqlc.CustomDomain) domain.CustomDomain {
	result := domain.CustomDomain{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Domain: row.Domain, Status: row.Status, VerificationToken: row.VerificationToken, SSLStatus: row.SslStatus}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		result.VerifiedAt = &t
	}
	result.CreatedAt = row.CreatedAt
	result.UpdatedAt = row.UpdatedAt
	return result
}
