package customersync

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Contact struct {
	Name       string
	Email      string
	Phone      string
	TotalSpent float64
	CountVisit bool
}

// UpsertCRMCustomer creates or updates a CRM customer scoped to businessID.
func UpsertCRMCustomer(ctx context.Context, db DBTX, businessID int64, contact Contact) (int64, error) {
	if businessID <= 0 {
		return 0, nil
	}

	name := strings.TrimSpace(contact.Name)
	email := strings.ToLower(strings.TrimSpace(contact.Email))
	phone := strings.TrimSpace(contact.Phone)
	if name == "" && email == "" && phone == "" {
		return 0, nil
	}

	firstName, lastName := splitCustomerName(name, email, phone)
	totalSpent := strconv.FormatFloat(contact.TotalSpent, 'f', 2, 64)
	visitCount := 0
	var lastVisit sql.NullTime
	if contact.CountVisit {
		visitCount = 1
		lastVisit = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}

	var customerID int64
	err := db.QueryRowContext(ctx,
		`SELECT id FROM crm_customers
		 WHERE business_id = ?
		   AND ((? <> '' AND LOWER(email) = LOWER(?)) OR (? <> '' AND phone = ?))
		 ORDER BY CASE WHEN ? <> '' AND LOWER(email) = LOWER(?) THEN 0 ELSE 1 END
		 LIMIT 1`,
		businessID, email, email, phone, phone, email, email).Scan(&customerID)
	if errors.Is(err, sql.ErrNoRows) {
		var ownerID int64
		if lookupErr := db.QueryRowContext(ctx,
			`SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE LIMIT 1`,
			businessID,
		).Scan(&ownerID); lookupErr != nil {
			return 0, lookupErr
		}

		result, insertErr := db.ExecContext(ctx,
			`INSERT INTO crm_customers
			 (owner_id, business_id, first_name, last_name, email, phone, total_spent, visit_count, last_visit_at, loyalty_tier, tags, is_active)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', JSON_ARRAY(), TRUE)`,
			ownerID,
			businessID,
			firstName,
			nullString(lastName),
			nullString(email),
			nullString(phone),
			totalSpent,
			visitCount,
			lastVisit,
		)
		if insertErr != nil {
			return 0, insertErr
		}
		return result.LastInsertId()
	}
	if err != nil {
		return 0, err
	}

	_, err = db.ExecContext(ctx,
		`UPDATE crm_customers
		 SET first_name = CASE WHEN ? <> '' THEN ? ELSE first_name END,
		     last_name = CASE WHEN ? <> '' THEN ? ELSE last_name END,
		     email = CASE WHEN ? <> '' THEN ? ELSE email END,
		     phone = CASE WHEN ? <> '' THEN ? ELSE phone END,
		     total_spent = total_spent + ?,
		     visit_count = visit_count + ?,
		     last_visit_at = CASE WHEN ? THEN ? ELSE last_visit_at END,
		     is_active = TRUE
		 WHERE business_id = ? AND id = ?`,
		firstName, firstName,
		lastName, lastName,
		email, email,
		phone, phone,
		totalSpent,
		visitCount,
		contact.CountVisit,
		lastVisit,
		businessID, customerID,
	)
	if err != nil {
		return 0, err
	}
	return customerID, nil
}

func splitCustomerName(name, email, phone string) (string, string) {
	if name == "" {
		if email != "" {
			name = strings.Split(email, "@")[0]
		} else if phone != "" {
			name = phone
		} else {
			name = "Customer"
		}
	}

	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "Customer", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}
