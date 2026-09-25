package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/staff/internal/domain"
	"github.com/voxire/onion/internal/modules/staff/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
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

// dynamic-query: squirrel
func (r *Repository) ListStaff(ctx context.Context, businessID int64, search, departmentID, status string) ([]domain.Staff, error) {
	// cross-schema read from hr_departments — to include department name in staff list response
	builder := sq.
		Select("s.id", "s.owner_id", "s.business_id", "s.user_id", "s.department_id",
			"COALESCE(d.name, '') AS department_name",
			"s.first_name", "s.last_name", "s.email", "s.phone",
			"s.role", "s.employment_type", "s.status", "s.salary", "s.hourly_rate", "s.overtime_rate",
			"s.start_date", "s.end_date", "s.address", "s.emergency_contact_name", "s.emergency_contact_phone",
			"s.apps", "s.notes", "s.photo_url", "s.bio", "s.created_at", "s.updated_at").
		From("hr_staff s").
		LeftJoin("hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id").
		Where(sq.Eq{"s.business_id": businessID})

	search = strings.TrimSpace(search)
	if search != "" {
		searchTerm := searchutil.BooleanPrefixQuery(search)
		builder = builder.Where(sq.Or{
			sq.Eq{"s.email": search},
			sq.Eq{"s.phone": search},
			sq.Expr("(? <> '' AND MATCH(s.first_name, s.last_name, s.email, s.role) AGAINST (? IN BOOLEAN MODE))", searchTerm, searchTerm),
		})
	}
	if departmentID != "" {
		builder = builder.Where(sq.Eq{"s.department_id": departmentID})
	}
	if status != "" {
		builder = builder.Where(sq.Eq{"s.status": status})
	}

	builder = builder.OrderBy("s.last_name ASC", "s.first_name ASC")

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list staff query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanStaffRows(rows)
}

// scanStaffRows reads the shared staff-list projection (s.id through
// s.updated_at, plus the joined department name) used by both ListStaff and
// ListStaffDirectory. Keeping the scan in one place means a column added to
// one caller's SELECT is scanned the same way for the other.
func scanStaffRows(rows *sql.Rows) ([]domain.Staff, error) {
	staff := make([]domain.Staff, 0)
	for rows.Next() {
		var s domain.Staff
		var (
			userID                sql.NullInt64
			departmentID_         sql.NullInt64
			departmentName        sql.NullString
			phone                 sql.NullString
			employmentType        sql.NullString
			statusVal             sql.NullString
			salary                sql.NullString
			hourlyRate            sql.NullString
			overtimeRate          sql.NullString
			endDate               sql.NullTime
			address               sql.NullString
			emergencyContactName  sql.NullString
			emergencyContactPhone sql.NullString
			apps                  []byte
			notes                 sql.NullString
			photoURL              sql.NullString
			bio                   sql.NullString
			createdAt             sql.NullTime
			updatedAt             sql.NullTime
		)
		if err := rows.Scan(
			&s.ID, &s.OwnerID, &s.BusinessID, &userID, &departmentID_, &departmentName,
			&s.FirstName, &s.LastName, &s.Email, &phone,
			&s.Role, &employmentType, &statusVal, &salary, &hourlyRate, &overtimeRate,
			&s.StartDate, &endDate, &address, &emergencyContactName, &emergencyContactPhone,
			&apps, &notes, &photoURL, &bio, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		if userID.Valid {
			s.UserID = userID.Int64
		}
		if departmentID_.Valid {
			s.DepartmentID = departmentID_.Int64
		}
		if departmentName.Valid {
			s.DepartmentName = departmentName.String
		}
		if photoURL.Valid {
			s.PhotoURL = photoURL.String
		}
		if bio.Valid {
			s.Bio = bio.String
		}
		if phone.Valid {
			s.Phone = phone.String
		}
		if employmentType.Valid {
			s.EmploymentType = employmentType.String
		}
		if statusVal.Valid {
			s.Status = statusVal.String
		}
		if salary.Valid {
			if v, err2 := strconv.ParseFloat(salary.String, 64); err2 == nil {
				s.Salary = v
			}
		}
		if hourlyRate.Valid {
			if v, err2 := strconv.ParseFloat(hourlyRate.String, 64); err2 == nil {
				s.HourlyRate = v
			}
		}
		if overtimeRate.Valid {
			if v, err2 := strconv.ParseFloat(overtimeRate.String, 64); err2 == nil {
				s.OvertimeRate = v
			}
		}
		if endDate.Valid {
			s.EndDate = endDate.Time
		}
		if address.Valid {
			s.Address = address.String
		}
		if emergencyContactName.Valid {
			s.EmergencyContactName = emergencyContactName.String
		}
		if emergencyContactPhone.Valid {
			s.EmergencyContactPhone = emergencyContactPhone.String
		}
		if apps != nil {
			s.Apps = string(apps)
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if createdAt.Valid {
			s.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			s.UpdatedAt = updatedAt.Time
		}
		staff = append(staff, s)
	}

	return staff, rows.Err()
}

// ListStaffDirectory returns one page of the business-wide staff list,
// ordered the same way as ListStaff. Unlike ListStaff it takes no search/
// department/status filters - the directory is deliberately the plain
// roster, not the admin screen's filtered view.
func (r *Repository) ListStaffDirectory(ctx context.Context, businessID, limit, offset int64) ([]domain.Staff, error) {
	query, args, err := sq.
		Select("s.id", "s.owner_id", "s.business_id", "s.user_id", "s.department_id",
			"COALESCE(d.name, '') AS department_name",
			"s.first_name", "s.last_name", "s.email", "s.phone",
			"s.role", "s.employment_type", "s.status", "s.salary", "s.hourly_rate", "s.overtime_rate",
			"s.start_date", "s.end_date", "s.address", "s.emergency_contact_name", "s.emergency_contact_phone",
			"s.apps", "s.notes", "s.photo_url", "s.bio", "s.created_at", "s.updated_at").
		From("hr_staff s").
		LeftJoin("hr_departments d ON d.id = s.department_id AND d.business_id = s.business_id").
		Where(sq.Eq{"s.business_id": businessID}).
		OrderBy("s.last_name ASC", "s.first_name ASC").
		Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build staff directory query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list staff directory", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	return scanStaffRows(rows)
}
