package infrastructure

import (
	"context"
	"database/sql"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
)

func (r *Repository) SearchSalesContacts(ctx context.Context, businessID int64, query string) ([]domain.SalesContact, error) {
	query = strings.TrimSpace(query)
	searchTerm := searchutil.BooleanPrefixQuery(query)

	var rows *sql.Rows
	var err error
	if query == "" {
		rows, err = r.pool.QueryContext(ctx, `
SELECT id, business_id, first_name, last_name, email, phone
FROM crm_customers
WHERE business_id = ?
ORDER BY first_name LIMIT 20`, businessID)
	} else {
		rows, err = r.pool.QueryContext(ctx, `
SELECT id, business_id, first_name, last_name, email, phone
FROM crm_customers
WHERE business_id = ?
  AND (
    phone = ?
    OR email = ?
    OR (? <> '' AND MATCH(first_name, last_name, email, phone) AGAINST (? IN BOOLEAN MODE))
  )
ORDER BY first_name LIMIT 20`, businessID, query, query, searchTerm, searchTerm)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to search POS customers", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	customers := make([]domain.SalesContact, 0)
	for rows.Next() {
		var row struct {
			ID         int64
			BusinessID int64
			FirstName  string
			LastName   sql.NullString
			Email      sql.NullString
			Phone      sql.NullString
		}
		if scanErr := rows.Scan(
			&row.ID,
			&row.BusinessID,
			&row.FirstName,
			&row.LastName,
			&row.Email,
			&row.Phone,
		); scanErr != nil {
			return nil, scanErr
		}

		customers = append(customers, domain.SalesContact{
			ID:        row.ID,
			FirstName: row.FirstName,
			LastName:  row.LastName.String,
			Email:     row.Email.String,
			Phone:     row.Phone.String,
		})
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}

	return customers, nil
}
