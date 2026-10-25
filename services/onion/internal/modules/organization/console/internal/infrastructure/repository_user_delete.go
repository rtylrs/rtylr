package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/organization/console/internal/application"
)

// userErasureStatements run in order once DeleteUser's guards pass. Access,
// credentials and personal data are deleted. HR and POS records keep their
// history without the link to the account, and rows the user authored in
// business data keep the bare id.
var userErasureStatements = []string{
	`UPDATE businesses SET is_active = FALSE WHERE owner_id = ? AND COALESCE(is_active, TRUE) = TRUE`,
	`UPDATE subscriptions SET status = 'cancelled'
		WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = ?) AND status IN ('active', 'trial', 'past_due')`,
	`UPDATE manual_access_grants SET revoked_at = NOW()
		WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = ?) AND revoked_at IS NULL`,
	`UPDATE manual_addon_grants SET revoked_at = NOW()
		WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = ?) AND revoked_at IS NULL`,
	`DELETE scope FROM authorization_role_assignment_scopes scope
		JOIN authorization_role_assignments assignment ON assignment.id = scope.assignment_id
		WHERE assignment.user_id = ?`,
	`DELETE FROM authorization_role_assignments WHERE user_id = ?`,
	`DELETE FROM authorization_license_assignments WHERE user_id = ?`,
	`DELETE FROM authorization_business_members WHERE user_id = ?`,
	`DELETE FROM user_group_members WHERE user_id = ?`,
	`DELETE FROM vox_user_permission_groups WHERE user_id = ?`,
	`DELETE FROM pos_member_locations WHERE user_id = ?`,
	`UPDATE pos_users SET is_active = FALSE WHERE user_id = ?`,
	`DELETE FROM flow_workspace_members WHERE user_id = ?`,
	`DELETE FROM flow_project_members WHERE user_id = ?`,
	`UPDATE hr_employees SET user_id = NULL WHERE user_id = ?`,
	`UPDATE hr_staff SET user_id = NULL WHERE user_id = ?`,
	`DELETE FROM user_sessions WHERE user_id = ?`,
	`DELETE FROM oauth_tokens WHERE user_id = ?`,
	`DELETE FROM oauth_auth_codes WHERE user_id = ?`,
	`DELETE FROM user_mfa_enrollments WHERE user_id = ?`,
	`DELETE FROM user_email_verification_tokens WHERE user_id = ?`,
	`DELETE FROM user_provisioning_states WHERE user_id = ?`,
	`DELETE FROM access_requests WHERE user_id = ?`,
	`DELETE FROM invite_tokens WHERE user_id = ?`,
	`DELETE FROM login_history WHERE user_id = ?`,
	`DELETE FROM hub_notifications WHERE user_id = ?`,
	`DELETE FROM flow_notifications WHERE user_id = ?`,
	`DELETE FROM flow_user_settings WHERE user_id = ?`,
	`DELETE FROM flow_task_focus_pins WHERE user_id = ?`,
	`DELETE FROM caret_reply_feedback WHERE user_id = ?`,
	`DELETE FROM conversation_messages WHERE user_id = ?`,
	`DELETE FROM conversation_summaries WHERE user_id = ?`,
	`DELETE FROM chat_sessions WHERE user_id = ?`,
	`DELETE FROM user_memory WHERE user_id = ?`,
	`DELETE FROM users WHERE id = ?`,
}

// DeleteUser erases a user in one transaction and deactivates the businesses
// only they use. It refuses while another member uses a business they own, or
// while their Paddle subscription can still bill.
func (r *Repository) DeleteUser(ctx context.Context, id int64) (*application.DeletedUser, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	rollback := func(cause error) (*application.DeletedUser, error) {
		_ = tx.Rollback()
		return nil, cause
	}

	var lockedID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id = ? FOR UPDATE`, id).Scan(&lockedID); err != nil {
		return rollback(err)
	}

	var billed int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM subscriptions
		WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = ?)
		  AND status IN ('active', 'trial', 'past_due')
		  AND COALESCE(paddle_subscription_id, '') <> ''`, id).Scan(&billed); err != nil {
		return rollback(err)
	}
	if billed > 0 {
		return rollback(application.ErrUserHasBilledPlan)
	}

	owned, err := queryBusinessRefs(ctx, tx, `
		SELECT id, name FROM businesses
		WHERE owner_id = ? AND COALESCE(is_active, TRUE) = TRUE
		ORDER BY id
		FOR UPDATE`, id)
	if err != nil {
		return rollback(err)
	}
	inUse, err := queryBusinessRefs(ctx, tx, `
		SELECT b.id, b.name FROM businesses b
		WHERE b.owner_id = ? AND COALESCE(b.is_active, TRUE) = TRUE
		  AND EXISTS (
		    SELECT 1 FROM authorization_business_members m
		    WHERE m.business_id = b.id AND m.user_id <> b.owner_id AND m.status IN ('active', 'suspended'))
		ORDER BY b.id`, id)
	if err != nil {
		return rollback(err)
	}
	if len(inUse) > 0 {
		return rollback(&application.OwnedBusinessesInUseError{Businesses: inUse})
	}

	deleted := &application.DeletedUser{}
	for _, business := range owned {
		deleted.DeactivatedBusinessIDs = append(deleted.DeactivatedBusinessIDs, business.ID)
	}
	if deleted.MemberBusinessIDs, err = queryIDs(ctx, tx, `
		SELECT DISTINCT m.business_id FROM authorization_business_members m
		JOIN businesses b ON b.id = m.business_id
		WHERE m.user_id = ? AND b.owner_id <> m.user_id
		ORDER BY m.business_id`, id); err != nil {
		return rollback(err)
	}
	if deleted.SeatBusinessIDs, err = queryIDs(ctx, tx, `
		SELECT DISTINCT m.business_id FROM authorization_business_members m
		JOIN businesses b ON b.id = m.business_id
		WHERE m.user_id = ? AND m.status = 'active' AND b.owner_id <> m.user_id
		  AND COALESCE(b.is_active, TRUE) = TRUE
		ORDER BY m.business_id`, id); err != nil {
		return rollback(err)
	}

	for _, statement := range userErasureStatements {
		if _, err := tx.ExecContext(ctx, statement, id); err != nil {
			return rollback(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return deleted, nil
}

func queryBusinessRefs(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]application.BusinessRef, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []application.BusinessRef
	for rows.Next() {
		var ref application.BusinessRef
		if err := rows.Scan(&ref.ID, &ref.Name); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func queryIDs(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
