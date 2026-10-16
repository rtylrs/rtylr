package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/infrastructure/persistence/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(db *sql.DB) *Repository { return &Repository{queries: sqlc.New(db)} }

func (r *Repository) Get(ctx context.Context, businessID int64) (domain.CustomDomain, error) {
	row, err := r.queries.GetCustomDomain(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CustomDomain{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CustomDomain{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) GetVerifiedByDomain(ctx context.Context, domainName string) (domain.CustomDomain, error) {
	row, err := r.queries.GetVerifiedCustomDomainByDomain(ctx, domainName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CustomDomain{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.CustomDomain{}, err
	}
	return fromRow(sqlc.GetCustomDomainRow(row)), nil
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// Register writes this business's pending claim, replacing a previous claim
// of its own (a re-register rotates the challenge, a new domain restarts it).
func (r *Repository) Register(ctx context.Context, input domain.CustomDomain) (domain.CustomDomain, error) {
	current, err := r.Get(ctx, input.BusinessID)
	switch {
	case err == nil && current.Domain == input.Domain:
		err = r.queries.RotateCustomDomainChallenge(ctx, sqlc.RotateCustomDomainChallengeParams{BusinessID: input.BusinessID, Domain: input.Domain, VerificationToken: input.VerificationToken})
	case err == nil:
		err = r.queries.ReplaceCustomDomain(ctx, sqlc.ReplaceCustomDomainParams{OwnerID: input.OwnerID, BusinessID: input.BusinessID, Domain: input.Domain, VerificationToken: input.VerificationToken})
	case errors.Is(err, domain.ErrNotFound):
		_, err = r.queries.RegisterCustomDomain(ctx, sqlc.RegisterCustomDomainParams{OwnerID: input.OwnerID, BusinessID: input.BusinessID, Domain: input.Domain, VerificationToken: input.VerificationToken})
	}
	if isDuplicate(err) {
		return domain.CustomDomain{}, domain.ErrConflict
	}
	if err != nil {
		return domain.CustomDomain{}, err
	}
	return r.Get(ctx, input.BusinessID)
}

// Verify marks the claim verified. The unique index on verified claims makes
// the first business to prove control win; competing pending claims are dropped.
func (r *Repository) Verify(ctx context.Context, businessID int64, token string) error {
	current, err := r.Get(ctx, businessID)
	if err != nil {
		return err
	}
	result, err := r.queries.VerifyCustomDomain(ctx, sqlc.VerifyCustomDomainParams{BusinessID: businessID, VerificationToken: token})
	if isDuplicate(err) {
		return domain.ErrConflict
	}
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrInvalidToken
	}
	return r.queries.DeleteCompetingCustomDomainClaims(ctx, sqlc.DeleteCompetingCustomDomainClaimsParams{Domain: current.Domain, BusinessID: businessID})
}

func (r *Repository) Remove(ctx context.Context, businessID int64) error {
	return r.queries.RemoveCustomDomain(ctx, businessID)
}

func fromRow(row sqlc.GetCustomDomainRow) domain.CustomDomain {
	result := domain.CustomDomain{ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Domain: row.Domain, Status: row.Status, VerificationToken: row.VerificationToken, SSLStatus: row.SslStatus}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		result.VerifiedAt = &t
	}
	result.CreatedAt = row.CreatedAt
	result.UpdatedAt = row.UpdatedAt
	return result
}
