package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) loadAuthorizationRoleAssignments(ctx context.Context, businessID, userID int64) ([]authorization.RoleAssignment, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT a.id, a.app_id, a.group_id, a.role_id, a.source, a.is_primary,
		       s.scope_type, s.scope_id
		FROM authorization_role_assignments a
		LEFT JOIN authorization_role_assignment_scopes s ON s.assignment_id = a.id
		WHERE a.business_id = ?
		  AND a.status = 'active'
		  AND (a.user_id = ? OR a.group_id IN (
			SELECT group_id
			FROM user_group_members
			WHERE business_id = ? AND user_id = ?
		  ))
		ORDER BY a.id`, businessID, userID, businessID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type assignmentRow struct {
		assignment authorization.RoleAssignment
	}
	byID := make(map[int64]*assignmentRow)
	order := make([]int64, 0)
	for rows.Next() {
		var (
			assignmentID int64
			appID        string
			groupID      sql.NullInt64
			roleID       int64
			source       string
			isPrimary    bool
			scopeType    sql.NullString
			scopeID      sql.NullString
		)
		if err := rows.Scan(&assignmentID, &appID, &groupID, &roleID, &source, &isPrimary, &scopeType, &scopeID); err != nil {
			return nil, err
		}
		entry, ok := byID[assignmentID]
		if !ok {
			entry = &assignmentRow{assignment: authorization.RoleAssignment{
				AppID:   authorization.AppID(appID),
				GroupID: nullInt64Pointer(groupID),
				Source:  authorization.AssignmentSource(source),
				RoleIDs: []int64{},
				Scopes:  []authorization.Scope{},
				Active:  true,
			}}
			byID[assignmentID] = entry
			order = append(order, assignmentID)
		}
		entry.assignment.RoleIDs = appendUniqueInt64(entry.assignment.RoleIDs, roleID)
		if isPrimary {
			primaryRoleID := roleID
			entry.assignment.PrimaryRoleID = &primaryRoleID
		}
		if scopeType.Valid && scopeID.Valid {
			entry.assignment.Scopes = appendUniqueAuthorizationScope(entry.assignment.Scopes, authorization.Scope{
				Type: scopeType.String,
				ID:   scopeID.String,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	assignments := make([]authorization.RoleAssignment, 0, len(order))
	for _, assignmentID := range order {
		assignments = append(assignments, byID[assignmentID].assignment)
	}
	return assignments, nil
}

func authorizationSchemaUnavailable(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "authorization_") &&
		(strings.Contains(message, "doesn't exist") || strings.Contains(message, "no such table"))
}

func authorizationAppPointer(appID string) *authorization.AppID {
	if appID == "" || appID == "dash" {
		return nil
	}
	value := authorization.AppID(appID)
	return &value
}

func nullInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func appendUniqueInt64(values []int64, value int64) []int64 {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueAuthorizationScope(values []authorization.Scope, value authorization.Scope) []authorization.Scope {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func (r *Repository) UpsertProvisioningState(ctx context.Context, userID, businessID int64, mode, state string, emailVerified, passwordChangeRequired bool) error {
	_, err := r.pool.ExecContext(ctx, `
		INSERT INTO user_provisioning_states
			(user_id, business_id, mode, state, email_verified, password_change_required)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			business_id = VALUES(business_id),
			mode = VALUES(mode),
			state = VALUES(state),
			email_verified = VALUES(email_verified),
			password_change_required = VALUES(password_change_required),
			updated_at = NOW()`,
		userID, businessID, mode, state, emailVerified, passwordChangeRequired,
	)
	if err != nil && authorizationSchemaUnavailable(err) {
		return nil
	}
	return err
}

func (r *Repository) GetProvisioningStatus(ctx context.Context, userID int64) (*authorization.ProvisioningStatus, bool, error) {
	var (
		mode, state                                 string
		emailVerified, passwordChangeRequired       bool
		mfaRequired, mfaEnrolled, recoveryAvailable bool
		lastMFAVerified                             sql.NullTime
	)
	err := r.pool.QueryRowContext(ctx, `
		SELECT mode, state, email_verified, password_change_required,
		       mfa_required, mfa_enrolled, recovery_available, last_mfa_verified_at
		FROM user_provisioning_states
		WHERE user_id = ?`, userID).Scan(
		&mode, &state, &emailVerified, &passwordChangeRequired,
		&mfaRequired, &mfaEnrolled, &recoveryAvailable, &lastMFAVerified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		if authorizationSchemaUnavailable(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	mfaVerified := !mfaRequired || (mfaEnrolled && lastMFAVerified.Valid && time.Since(lastMFAVerified.Time) <= 5*time.Minute)
	return &authorization.ProvisioningStatus{
		Mode:                   authorization.ProvisioningMode(mode),
		State:                  authorization.ProvisioningState(state),
		EmailVerified:          emailVerified,
		PasswordChangeRequired: passwordChangeRequired,
		MFARequired:            mfaRequired,
		MFAEnrolled:            mfaEnrolled,
		MFAVerified:            mfaVerified,
		RecoveryAvailable:      recoveryAvailable,
	}, true, nil
}

func (r *Repository) IsMFAVerified(ctx context.Context, userID int64) (bool, error) {
	var required, enrolled bool
	var lastVerified sql.NullTime
	err := r.pool.QueryRowContext(ctx, `
		SELECT mfa_required, mfa_enrolled, last_mfa_verified_at
		FROM user_provisioning_states
		WHERE user_id = ?`, userID).Scan(&required, &enrolled, &lastVerified)
	if errors.Is(err, sql.ErrNoRows) || authorizationSchemaUnavailable(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !required {
		return true, nil
	}
	return enrolled && lastVerified.Valid && time.Since(lastVerified.Time) <= 5*time.Minute, nil
}

func (r *Repository) MarkPasswordChanged(ctx context.Context, userID int64) error {
	var businessID int64
	if err := r.pool.QueryRowContext(ctx, `
		SELECT business_id FROM user_provisioning_states WHERE user_id = ?`, userID).Scan(&businessID); err != nil {
		if errors.Is(err, sql.ErrNoRows) || authorizationSchemaUnavailable(err) {
			return nil
		}
		return err
	}
	_, err := r.pool.ExecContext(ctx, `
		UPDATE user_provisioning_states
		SET password_change_required = FALSE,
		    state = CASE WHEN email_verified THEN 'active' ELSE 'pending_activation' END,
		    password_changed_at = NOW(),
		    updated_at = NOW()
		WHERE user_id = ?`, userID)
	if err != nil && authorizationSchemaUnavailable(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.ActivateNormalizedAuthorization(ctx, businessID, userID)
}

func (r *Repository) SetNormalizedAuthorizationPending(ctx context.Context, businessID, userID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		if authorizationSchemaUnavailable(cause) {
			return nil
		}
		return cause
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'pending', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE business_id = ? AND user_id = ?`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'pending'
		WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'pending'
		WHERE business_id = ? AND user_id = ? AND status = 'active'`, businessID, userID); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

func (r *Repository) ActivateNormalizedAuthorization(ctx context.Context, businessID, userID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		_ = tx.Rollback()
		if authorizationSchemaUnavailable(cause) {
			return nil
		}
		return cause
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_business_members
		SET status = 'active', authorization_version = authorization_version + 1, updated_at = NOW()
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_license_assignments
		SET status = 'active'
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, businessID, userID); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE authorization_role_assignments
		SET status = 'active'
		WHERE business_id = ? AND user_id = ? AND status = 'pending'`, businessID, userID); err != nil {
		return rollback(err)
	}
	return tx.Commit()
}

// GetOwnerSubscriptionInfo returns the plan slug and status for the owner's subscription.
func (r *Repository) GetOwnerSubscriptionInfo(ctx context.Context, ownerID int64) (string, string, error) {
	// cross-schema read from subscriptions + plans — approved for owner subscription info
	row, err := r.queries.GetOwnerSubscriptionInfo(ctx, ownerID)
	if err != nil {
		return "", "", err
	}
	status := ""
	if row.Status.Valid {
		status = string(row.Status.SubscriptionsStatus)
	}
	return row.Slug, status, nil
}

// RecordSeatChange updates the billing seat_periods table after a team member is added or removed.
// All three writes (close the active period, open a new one, update the subscription seat count)
// run inside a single transaction so billing state can never end up partially applied.
func (r *Repository) RecordSeatChange(ctx context.Context, ownerID int64, newCount int) error {
	now := time.Now().UTC()
	yesterday := now.AddDate(0, 0, -1)
	billingMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	// cross-schema read from subscriptions — approved for seat period management
	sub, err := qtx.LockSubscriptionForSeatChange(ctx, ownerID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	if err := qtx.CloseActiveSeatPeriod(ctx, sqlc.CloseActiveSeatPeriodParams{
		PeriodEnd:      sql.NullTime{Time: yesterday, Valid: true},
		SubscriptionID: sub.ID,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to seat_periods — approved for seat billing accuracy
	if err := qtx.OpenSeatPeriod(ctx, sqlc.OpenSeatPeriodParams{
		SubscriptionID:    sub.ID,
		OwnerID:           ownerID,
		SeatCount:         int32(newCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PricePerSeatCents: sub.PricePerSeatCents,
		PeriodStart:       now,
		BillingMonth:      billingMonth,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	// cross-schema write to subscriptions — approved for seat billing accuracy
	if err := qtx.UpdateSubscriptionSeatCount(ctx, sqlc.UpdateSubscriptionSeatCountParams{
		SeatCount: int32(newCount), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		OwnerID:   ownerID,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ─── Access Requests ──────────────────────────────────────────────────────────

func (r *Repository) ListAccessRequests(ctx context.Context, businessID int64) ([]domain.AccessRequest, error) {
	rows, err := r.queries.ListAccessRequestsByBusiness(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AccessRequest, 0, len(rows))
	for _, row := range rows {
		status := "pending"
		if row.Status.Valid {
			status = string(row.Status.AccessRequestsStatus)
		}
		out = append(out, domain.AccessRequest{
			ID:             row.ID,
			UserID:         row.UserID,
			App:            row.App,
			Status:         status,
			RequesterName:  row.RequesterName,
			RequesterEmail: row.RequesterEmail,
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return out, nil
}

// GetAdminsAndOwnerUserIDs returns all user IDs that should receive management
// notifications for a business: the primary owner plus any staff with
// role in ('admin','owner') in user_app_permissions.
func (r *Repository) GetAdminsAndOwnerUserIDs(ctx context.Context, businessID int64) ([]int64, error) {
	// cross-schema read from hub — approved for business-wide admin notification fan-out
	return r.queries.ListAdminsAndOwnerUserIDs(ctx, sqlc.ListAdminsAndOwnerUserIDsParams{
		ID:         businessID,
		BusinessID: sql.NullInt64{Int64: businessID, Valid: true},
	})
}

// BulkCreateHubNotification writes an in-app notification to hub_notifications for
// each recipient in userIDs. Best-effort: errors are logged, not propagated.
func (r *Repository) ResolveAccessRequest(ctx context.Context, id, businessID int64, status string) error {
	return r.queries.ResolveAccessRequest(ctx, sqlc.ResolveAccessRequestParams{
		ID:         id,
		BusinessID: businessID,
		Status: sqlc.NullAccessRequestsStatus{
			AccessRequestsStatus: sqlc.AccessRequestsStatus(status),
			Valid:                true,
		},
	})
}

func (r *Repository) BulkCreateHubNotification(ctx context.Context, ownerID, businessID int64, userIDs []int64, title, body, notifType, link string) error {
	for _, uid := range userIDs {
		// cross-schema write to hub_notifications — approved for cross-module access request alerts
		if err := r.queries.CreateHubNotification(ctx, sqlc.CreateHubNotificationParams{
			OwnerID:    ownerID,
			BusinessID: businessID,
			UserID:     uid,
			Title:      title,
			Body:       sql.NullString{String: body, Valid: body != ""},
			Type:       notifType,
			Link:       sql.NullString{String: link, Valid: link != ""},
		}); err != nil {
			logger.Ctx(ctx).Warn("failed to create hub notification",
				logger.Int("user_id", int(uid)), logger.Error(err))
		}
	}
	return nil
}

// nullBusinessID wraps the join-side business id so sqlc's named parameter
// (shared by both placeholders in RevokeBusinessUserPermissions) scans as
// NOT NULL int64.
func nullBusinessID(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
