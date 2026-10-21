package application

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/organization/team"
	"github.com/voxire/onion/internal/modules/webhooks/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/eventbus"
)

func newTestSender(db *sql.DB) *Sender {
	return NewSender(infrastructure.NewRepository(db), eventbus.NewSQLReader(db))
}

func newTestNotifier(db *sql.DB) *Notifier {
	return NewNotifier(
		infrastructure.NewRepository(db),
		testEmailQueue{db: db},
		team.NewModule(context.Background(), db).GetActiveMemberContact,
	)
}

type testEmailQueue struct{ db *sql.DB }

func (q testEmailQueue) Enqueue(ctx context.Context, message NotificationEmail) error {
	return emailqueue.Enqueue(ctx, q.db, emailqueue.Job{
		To:        message.To,
		Subject:   message.Subject,
		HTML:      message.HTML,
		Text:      message.Text,
		DedupeKey: message.DedupeKey,
	})
}
