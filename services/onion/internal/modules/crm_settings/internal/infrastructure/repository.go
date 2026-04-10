package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) Get(ctx context.Context, ownerID int64) (domain.Settings, error) {
	var s domain.Settings
	var fromEmail, fromName, smsSenderID, timezone sql.NullString

	err := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, default_consent, from_email, from_name, sms_sender_id, timezone, created_at, updated_at
		 FROM crm_settings WHERE owner_id = ?`, ownerID,
	).Scan(&s.ID, &s.OwnerID, &s.DefaultConsent, &fromEmail, &fromName, &smsSenderID, &timezone, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Settings{
				OwnerID:        ownerID,
				DefaultConsent: true,
				Timezone:       "UTC",
			}, nil
		}

		logger.Ctx(ctx).Error("Failed to get CRM settings", logger.Error(err))

		return domain.Settings{}, err
	}

	s.FromEmail = fromEmail.String
	s.FromName = fromName.String
	s.SmsSenderID = smsSenderID.String
	s.Timezone = timezone.String

	if s.Timezone == "" {
		s.Timezone = "UTC"
	}

	return s, nil
}

func (r *Repository) Upsert(ctx context.Context, input domain.SettingsInput) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_settings
		   (owner_id, default_consent, from_email, from_name, sms_sender_id, timezone)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   default_consent = VALUES(default_consent),
		   from_email = VALUES(from_email),
		   from_name = VALUES(from_name),
		   sms_sender_id = VALUES(sms_sender_id),
		   timezone = VALUES(timezone)`,
		input.OwnerID, input.DefaultConsent, input.FromEmail,
		input.FromName, input.SmsSenderID, input.Timezone,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert CRM settings", logger.Error(err))
	}

	return err
}
