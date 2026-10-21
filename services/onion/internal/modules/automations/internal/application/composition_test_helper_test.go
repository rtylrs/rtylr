package application

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/eventbus"
)

func newTestExecutor(db *sql.DB) *Executor {
	return NewExecutor(infrastructure.NewRepository(db), testEmailQueue{db: db}, eventbus.NewSQLReader(db), testCustomerContactResolver(db))
}

func testCustomerContactResolver(db *sql.DB) func(context.Context, int64, int64) (domain.CustomerContact, bool, error) {
	return func(ctx context.Context, customerID, businessID int64) (domain.CustomerContact, bool, error) {
		var contact domain.CustomerContact
		err := db.QueryRowContext(ctx, `SELECT id, COALESCE(email, ''), TRIM(CONCAT(first_name, ' ', COALESCE(last_name, ''))) FROM crm_customers WHERE id = ? AND business_id = ?`, customerID, businessID).Scan(&contact.ID, &contact.Email, &contact.Name)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.CustomerContact{}, false, nil
		}
		return contact, err == nil, err
	}
}

type testEmailQueue struct{ db *sql.DB }

func (q testEmailQueue) Enqueue(ctx context.Context, message EmailMessage) error {
	return emailqueue.Enqueue(ctx, q.db, emailqueue.Job{
		To:        message.To,
		Subject:   message.Subject,
		HTML:      message.HTML,
		Text:      message.Text,
		DedupeKey: message.DedupeKey,
	})
}
