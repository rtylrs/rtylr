package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"github.com/voxire/onion/internal/modules/crm/internal/outbox"
	"strconv"
	"strings"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) CaptureCustomerContact(ctx context.Context, event eventbus.Event, ownerID int64, contact crmcontract.CustomerContactCaptured) (err error) {
	name := strings.TrimSpace(contact.Name)
	email := strings.ToLower(strings.TrimSpace(contact.Email))
	phone := strings.TrimSpace(contact.Phone)
	if email == "" && phone == "" {
		return nil
	}

	firstName, lastName := splitCapturedName(name, email, phone)
	totalSpent := strconv.FormatFloat(contact.TotalSpent, 'f', 2, 64)
	visitCount := int32(0)
	var lastVisit sql.NullTime
	if contact.CountVisit {
		visitCount = 1
		lastVisit = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	accepted, err := qtx.RecordCustomerCaptureEvent(ctx, sqlc.RecordCustomerCaptureEventParams{EventID: event.ID, BusinessID: event.BusinessID})
	if err != nil {
		return err
	}
	rows, err := accepted.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return tx.Commit()
	}

	customerID, err := qtx.FindCustomerForCapture(ctx, sqlc.FindCustomerForCaptureParams{
		BusinessID: event.BusinessID,
		Email:      email,
		PhoneInput: sqlutil.NullString(phone),
	})
	if errors.Is(err, sql.ErrNoRows) {
		result, insertErr := qtx.InsertCustomerFromCapture(ctx, sqlc.InsertCustomerFromCaptureParams{
			OwnerID:     ownerID,
			BusinessID:  event.BusinessID,
			FirstName:   firstName,
			LastName:    sqlutil.NullString(lastName),
			Email:       sqlutil.NullString(email),
			Phone:       sqlutil.NullString(phone),
			TotalSpent:  sql.NullString{String: totalSpent, Valid: true},
			VisitCount:  sql.NullInt32{Int32: visitCount, Valid: true},
			LastVisitAt: lastVisit,
		})
		if insertErr != nil {
			return insertErr
		}
		customerID, err = result.LastInsertId()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		updateFirstName := firstName
		updateLastName := sqlutil.NullString(lastName)
		if name == "" {
			updateFirstName = ""
			updateLastName = sqlutil.NullString("")
		}
		if err = qtx.UpdateCustomerFromCapture(ctx, sqlc.UpdateCustomerFromCaptureParams{
			FirstName:       updateFirstName,
			LastName:        updateLastName,
			Email:           sqlutil.NullString(email),
			Phone:           sqlutil.NullString(phone),
			TotalSpentDelta: sql.NullString{String: totalSpent, Valid: true},
			VisitCountDelta: sql.NullInt32{Int32: visitCount, Valid: true},
			LastVisitAt:     lastVisit,
			BusinessID:      event.BusinessID,
			CustomerID:      customerID,
		}); err != nil {
			return err
		}
	}
	linked := crmcontract.CustomerLinked{
		SourceType: event.SubjectType,
		SourceID:   event.SubjectID,
		CustomerID: customerID,
	}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  event.BusinessID,
		Type:        crmcontract.CustomerLinkedType,
		SubjectType: "crm_customer",
		SubjectID:   customerID,
		Payload:     linked,
		DedupeKey:   crmcontract.CustomerLinkDedupeKey(event.SubjectType, event.SubjectID, customerID),
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func splitCapturedName(name, email, phone string) (string, string) {
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
	switch len(parts) {
	case 0:
		return "Customer", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}
