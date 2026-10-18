package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/segmentrules"
)

// sendableStatuses are the states a campaign can be sent from.
var sendableStatuses = []string{"draft", "scheduled", "active", "paused"}

// ClaimCampaignForSend marks an email campaign sent and returns it, in one
// conditional update so two clicks cannot both send it. A campaign that is
// not email, or is already sent or cancelled, is ErrCampaignNotSendable.
func (r *Repository) ClaimCampaignForSend(ctx context.Context, businessID, campaignID int64) (domain.CampaignToSend, error) {
	claimSQL, claimArgs, err := sq.Update("crm_campaigns").
		Set("status", "sent").
		Set("sent_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": campaignID, "business_id": businessID, "channel": "email", "status": sendableStatuses}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.CampaignToSend{}, err
	}
	result, err := r.pool.ExecContext(ctx, claimSQL, claimArgs...)
	if err != nil {
		return domain.CampaignToSend{}, err
	}
	claimed, err := result.RowsAffected()
	if err != nil {
		return domain.CampaignToSend{}, err
	}

	var (
		campaign  domain.CampaignToSend
		subject   sql.NullString
		segmentID sql.NullInt64
	)
	err = r.pool.QueryRowContext(ctx,
		"SELECT id, name, subject, content, segment_id FROM crm_campaigns WHERE id = ? AND business_id = ?",
		campaignID, businessID,
	).Scan(&campaign.ID, &campaign.Name, &subject, &campaign.Content, &segmentID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CampaignToSend{}, domain.ErrCampaignNotFound
	}
	if err != nil {
		return domain.CampaignToSend{}, err
	}
	if claimed == 0 {
		return domain.CampaignToSend{}, domain.ErrCampaignNotSendable
	}
	campaign.Subject = subject.String
	if segmentID.Valid && segmentID.Int64 > 0 {
		campaign.SegmentID = &segmentID.Int64
	}
	return campaign, nil
}

// ReleaseCampaignClaim returns a claimed campaign to draft when sending
// fails before anything was queued, so it can be sent again.
func (r *Repository) ReleaseCampaignClaim(ctx context.Context, businessID, campaignID int64) error {
	_, err := r.pool.ExecContext(ctx,
		"UPDATE crm_campaigns SET status = 'draft', sent_at = NULL WHERE id = ? AND business_id = ? AND status = 'sent'",
		campaignID, businessID,
	)
	return err
}

// CampaignAudience lists the customers a campaign reaches: active ones in
// the segment (everyone when segmentID is nil) who have an email address and
// gave marketing consent. total counts the segment before that filter.
func (r *Repository) CampaignAudience(ctx context.Context, businessID int64, segmentID *int64) ([]domain.AudienceMember, int, error) {
	audience := sq.And{sq.Eq{"business_id": businessID}, sq.Eq{"is_active": true}}
	if segmentID != nil {
		segment, err := r.GetSegmentByID(ctx, businessID, *segmentID)
		if err != nil {
			return nil, 0, err
		}
		if frag, args := segmentrules.Filter(segment.Rules); frag != "" {
			audience = append(audience, sq.Expr(frag, args...))
		}
	}

	countSQL, countArgs, err := sq.Select("COUNT(*)").From("crm_customers").Where(audience).
		PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	eligible := append(sq.And{}, audience...)
	eligible = append(eligible,
		sq.Expr("COALESCE(email, '') <> ''"),
		// Consent is opt-in: a missing record or a string "true" is not consent.
		sq.Expr("JSON_EXTRACT(consent, '$.marketing') = CAST('true' AS JSON)"),
	)
	listSQL, listArgs, err := sq.Select("id", "first_name", "last_name", "email", "loyalty_points", "COALESCE(loyalty_tier, '')").
		From("crm_customers").Where(eligible).OrderBy("id").
		PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var members []domain.AudienceMember
	for rows.Next() {
		var (
			member    domain.AudienceMember
			firstName sql.NullString
			lastName  sql.NullString
			points    sql.NullInt64
		)
		if err := rows.Scan(&member.CustomerID, &firstName, &lastName, &member.Email, &points, &member.Tier); err != nil {
			return nil, 0, err
		}
		member.FirstName, member.LastName = firstName.String, lastName.String
		member.Email = strings.TrimSpace(member.Email)
		member.Points = int(points.Int64)
		members = append(members, member)
	}
	return members, total, rows.Err()
}

// RecordCampaignSends logs one queued send per recipient and the campaign's
// audience and sent counts.
func (r *Repository) RecordCampaignSends(ctx context.Context, businessID, campaignID int64, customerIDs []int64, audience int) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for start := 0; start < len(customerIDs); start += 500 {
		end := min(start+500, len(customerIDs))
		insert := sq.Insert("crm_campaign_sends").Columns("campaign_id", "business_id", "customer_id", "status", "sent_at")
		for _, customerID := range customerIDs[start:end] {
			insert = insert.Values(campaignID, businessID, customerID, "queued", sq.Expr("NOW()"))
		}
		insertSQL, insertArgs, err := insert.PlaceholderFormat(sq.Question).ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, insertSQL, insertArgs...); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE crm_campaigns SET audience_count = ?, sent_count = ? WHERE id = ? AND business_id = ?",
		audience, len(customerIDs), campaignID, businessID,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// UnsubscribeFromMarketing withdraws a customer's marketing consent, leaving
// the rest of their consent record as it was.
func (r *Repository) UnsubscribeFromMarketing(ctx context.Context, businessID, customerID int64) error {
	_, err := r.pool.ExecContext(ctx,
		"UPDATE crm_customers SET consent = JSON_SET(COALESCE(consent, JSON_OBJECT()), '$.marketing', CAST('false' AS JSON)) WHERE id = ? AND business_id = ?",
		customerID, businessID,
	)
	return err
}
