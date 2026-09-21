package customersync

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	customersyncsqlc "github.com/voxire/onion/internal/platform/customer_sync/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
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
	visitCount := int32(0)
	var lastVisit sql.NullTime
	if contact.CountVisit {
		visitCount = 1
		lastVisit = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}

	queries := customersyncsqlc.New(db)
	customerID, err := queries.FindCRMCustomerByContact(ctx, customersyncsqlc.FindCRMCustomerByContactParams{
		BusinessID: businessID,
		Email:      email,
		PhoneInput: sqlutil.NullString(phone),
	})
	if errors.Is(err, sql.ErrNoRows) {
		ownerID, lookupErr := queries.GetActiveBusinessOwnerID(ctx, businessID)
		if lookupErr != nil {
			return 0, lookupErr
		}

		result, insertErr := queries.InsertCRMCustomerFromSync(ctx, customersyncsqlc.InsertCRMCustomerFromSyncParams{
			OwnerID:     ownerID,
			BusinessID:  businessID,
			FirstName:   firstName,
			LastName:    sqlutil.NullString(lastName),
			Email:       sqlutil.NullString(email),
			Phone:       sqlutil.NullString(phone),
			TotalSpent:  sql.NullString{String: totalSpent, Valid: true},
			VisitCount:  sql.NullInt32{Int32: visitCount, Valid: true},
			LastVisitAt: lastVisit,
		})
		if insertErr != nil {
			return 0, insertErr
		}
		return result.LastInsertId()
	}
	if err != nil {
		return 0, err
	}

	updateLastVisit := sql.NullTime{}
	if contact.CountVisit {
		updateLastVisit = lastVisit
	}

	// The name fallback derived from the email/phone is only for fresh rows.
	// When the contact carried no name, keep the stored name instead of
	// overwriting it with the email prefix on every sync.
	updateFirstName := firstName
	updateLastName := sqlutil.NullString(lastName)
	if name == "" {
		updateFirstName = ""
		updateLastName = sql.NullString{}
	}

	err = queries.UpdateCRMCustomerFromSync(ctx, customersyncsqlc.UpdateCRMCustomerFromSyncParams{
		FirstName:       updateFirstName,
		LastName:        updateLastName,
		Email:           sqlutil.NullString(email),
		Phone:           sqlutil.NullString(phone),
		TotalSpentDelta: sql.NullString{String: totalSpent, Valid: true},
		VisitCountDelta: sql.NullInt32{Int32: visitCount, Valid: true},
		LastVisitAt:     updateLastVisit,
		BusinessID:      businessID,
		CustomerID:      customerID,
	})
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
