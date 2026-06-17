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

func UpsertCRMCustomerForBusiness(ctx context.Context, db DBTX, businessID int64, contact Contact) (int64, error) {
	if businessID <= 0 {
		return 0, nil
	}

	var ownerID int64
	if err := db.QueryRowContext(ctx, `SELECT owner_id FROM businesses WHERE id = ? LIMIT 1`, businessID).Scan(&ownerID); err != nil {
		return 0, err
	}

	return UpsertCRMCustomer(ctx, db, ownerID, contact)
}

func UpsertCRMCustomer(ctx context.Context, db DBTX, ownerID int64, contact Contact) (int64, error) {
	if ownerID <= 0 {
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
		 WHERE owner_id = ?
		   AND ((? <> '' AND LOWER(email) = LOWER(?)) OR (? <> '' AND phone = ?))
		 ORDER BY CASE WHEN ? <> '' AND LOWER(email) = LOWER(?) THEN 0 ELSE 1 END
		 LIMIT 1`,
		ownerID, email, email, phone, phone, email, email).Scan(&customerID)
	if errors.Is(err, sql.ErrNoRows) {
		result, insertErr := db.ExecContext(ctx,
			`INSERT INTO crm_customers
			 (owner_id, first_name, last_name, email, phone, total_spent, visit_count, last_visit_at, loyalty_tier, tags, is_active)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'bronze', JSON_ARRAY(), TRUE)`,
			ownerID,
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
		 WHERE owner_id = ? AND id = ?`,
		firstName, firstName,
		lastName, lastName,
		email, email,
		phone, phone,
		totalSpent,
		visitCount,
		contact.CountVisit,
		lastVisit,
		ownerID, customerID,
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
