package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	db           *sql.DB
	queries      *sqlc.Queries
	memberReader func(context.Context, int64, int64) (bool, error)
	repReader    func(context.Context, int64) ([]domain.Rep, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
}

func (r *Repository) WithMembershipReader(reader func(context.Context, int64, int64) (bool, error)) *Repository {
	r.memberReader = reader
	return r
}

func (r *Repository) WithRepReader(reader func(context.Context, int64) ([]domain.Rep, error)) *Repository {
	r.repReader = reader
	return r
}

// ValidateAssignee checks a user can be given work in the business.
func (r *Repository) ValidateAssignee(ctx context.Context, businessID, userID int64) error {
	if r.memberReader == nil {
		return errors.New("crm sales membership reader is not configured")
	}
	ok, err := r.memberReader(ctx, businessID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrInvalidAssignee
	}
	return nil
}

func (r *Repository) ListReps(ctx context.Context, businessID int64) ([]domain.Rep, error) {
	if r.repReader == nil {
		return nil, errors.New("crm sales rep reader is not configured")
	}
	return r.repReader(ctx, businessID)
}

// inTx runs fn in one transaction, committing only when it succeeds.
func (r *Repository) inTx(ctx context.Context, fn func(q *sqlc.Queries, tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to roll back crm sales transaction", logger.Error(rollbackErr))
		}
	}()
	if err := fn(r.queries.WithTx(tx), tx); err != nil {
		return err
	}
	return tx.Commit()
}

func notFound(err, target error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return target
	}
	return err
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func someInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: true}
}

func nullInt64s(values []int64) []sql.NullInt64 {
	out := make([]sql.NullInt64, 0, len(values))
	for _, value := range values {
		out = append(out, someInt64(value))
	}
	return out
}

func nullInt32(value *int32) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *value, Valid: true}
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func someTime(value time.Time) sql.NullTime {
	return sql.NullTime{Time: value, Valid: true}
}

// nullDate parses a YYYY-MM-DD date; "" is NULL. Callers validate first.
func nullDate(value string) sql.NullTime {
	if value == "" {
		return sql.NullTime{}
	}
	day, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: day, Valid: true}
}

func nullDatePtr(value *string) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return nullDate(*value)
}

func nullAmount(value *money.Amount) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func int64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time.UTC()
	return &v
}

func dateString(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.DateOnly)
}

func datePtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	v := value.Time.Format(time.DateOnly)
	return &v
}

// amount reads a DECIMAL column. The database only holds valid decimals, so a
// parse failure is logged rather than failing the whole read.
func amount(value string) money.Amount {
	parsed, err := money.FromString(value)
	if err != nil {
		logger.Global().Error("crm sales read an unparseable amount", logger.Any("value", value))
		return 0
	}
	return parsed
}

func amountPtr(value sql.NullString) *money.Amount {
	if !value.Valid || value.String == "" {
		return nil
	}
	v := amount(value.String)
	return &v
}

// likePattern turns search text into a contains pattern with LIKE's wildcards
// escaped.
func likePattern(search string) sql.NullString {
	search = strings.TrimSpace(search)
	if search == "" {
		return sql.NullString{}
	}
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(search)
	return sql.NullString{String: "%" + escaped + "%", Valid: true}
}

// digitsPattern matches a search against phone numbers once it holds enough
// digits to mean one.
func digitsPattern(search string) sql.NullString {
	digits := domain.PhoneDigits(search)
	if len(digits) < 3 {
		return sql.NullString{}
	}
	return sql.NullString{String: "%" + digits + "%", Valid: true}
}

func dayStart(day time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
}
