package emailqueue

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/testutil"
)

func TestPoolSendsDecryptedJobThroughConfiguredSender(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	db := testutil.OpenDB(t)
	ctx := context.Background()
	subject := "sender-" + uuid.NewString()
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE subject = ?", subject) })

	if err := Enqueue(ctx, db, Job{
		Priority:  127,
		To:        "queue-sender@test.local",
		Subject:   subject,
		HTML:      "<p>hello</p>",
		Text:      "hello",
		DedupeKey: subject,
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	var sent Job
	pool := NewPool(db, func(_ context.Context, job Job) error {
		sent = job
		return nil
	})
	if !pool.processOne(ctx) {
		t.Fatal("processOne returned false for a queued job")
	}
	if sent.To != "queue-sender@test.local" || sent.Subject != subject || sent.HTML != "<p>hello</p>" || sent.Text != "hello" {
		t.Fatalf("sender received %+v, want the decrypted queue job", sent)
	}

	var status string
	if err := db.QueryRowContext(ctx, "SELECT status FROM email_jobs WHERE subject = ?", subject).Scan(&status); err != nil {
		t.Fatalf("read queue status: %v", err)
	}
	if status != "sent" {
		t.Fatalf("queue status = %q, want sent", status)
	}
}
