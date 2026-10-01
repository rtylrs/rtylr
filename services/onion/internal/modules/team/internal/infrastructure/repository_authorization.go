package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
)

// GetOwnerApps returns the set of apps included in the business subscription.
func (r *Repository) GetOwnerApps(ctx context.Context, ownerID int64) ([]string, error) {
	appsJSON, err := r.queries.GetOwnerApps(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	var apps []string
	if err := json.Unmarshal(appsJSON, &apps); err != nil {
		return []string{}, nil
	}
	return apps, nil
}

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
	return err
}

func (r *Repository) GetProvisioningStatus(ctx context.Context, userID int64) (*authorization.ProvisioningStatus, bool, error) {
	return r.getProvisioningStatus(ctx, 0, userID)
}

func (r *Repository) GetProvisioningStatusForBusiness(ctx context.Context, businessID, userID int64) (*authorization.ProvisioningStatus, bool, error) {
	return r.getProvisioningStatus(ctx, businessID, userID)
}

func (r *Repository) getProvisioningStatus(ctx context.Context, businessID, userID int64) (*authorization.ProvisioningStatus, bool, error) {
	var (
		mode, state                                 string
		emailVerified, passwordChangeRequired       bool
		mfaRequired, mfaEnrolled, recoveryAvailable bool
		lastMFAVerified                             sql.NullTime
	)
	err := r.pool.QueryRowContext(ctx, `
		SELECT provisioning.mode, provisioning.state, provisioning.email_verified, provisioning.password_change_required,
		       (provisioning.mfa_required OR COALESCE(business.mfa_enforced, FALSE)),
		       provisioning.mfa_enrolled, provisioning.recovery_available, provisioning.last_mfa_verified_at
		FROM user_provisioning_states AS provisioning
		LEFT JOIN businesses AS business ON business.id = CASE WHEN ? > 0 THEN ? ELSE provisioning.business_id END
		WHERE provisioning.user_id = ?`, businessID, businessID, userID).Scan(
		&mode, &state, &emailVerified, &passwordChangeRequired,
		&mfaRequired, &mfaEnrolled, &recoveryAvailable, &lastMFAVerified,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
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
	return r.isMFAVerified(ctx, 0, userID)
}

func (r *Repository) IsMFAVerifiedForBusiness(ctx context.Context, businessID, userID int64) (bool, error) {
	return r.isMFAVerified(ctx, businessID, userID)
}

func (r *Repository) isMFAVerified(ctx context.Context, businessID, userID int64) (bool, error) {
	var required, enrolled bool
	var lastVerified sql.NullTime
	err := r.pool.QueryRowContext(ctx, `
		SELECT (provisioning.mfa_required OR COALESCE(business.mfa_enforced, FALSE)),
		       provisioning.mfa_enrolled, provisioning.last_mfa_verified_at
		FROM user_provisioning_states AS provisioning
		LEFT JOIN businesses AS business ON business.id = CASE WHEN ? > 0 THEN ? ELSE provisioning.business_id END
		WHERE provisioning.user_id = ?`, businessID, businessID, userID).Scan(&required, &enrolled, &lastVerified)
	if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
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
// notifications for a business: the primary owner plus active admins.
func (r *Repository) GetAdminsAndOwnerUserIDs(ctx context.Context, businessID int64) ([]int64, error) {
	// cross-schema read from hub — approved for business-wide admin notification fan-out
	return r.queries.ListAdminsAndOwnerUserIDs(ctx, sqlc.ListAdminsAndOwnerUserIDsParams{
		ID:         businessID,
		BusinessID: businessID,
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

// GetAuthorizationSnapshot loads the member's normalized authorization rows
// for the pure resolver.
func (r *Repository) GetAuthorizationSnapshot(ctx context.Context, businessID, userID int64, businessApps []string) (authorization.Snapshot, bool, error) {
	if businessID <= 0 || userID <= 0 {
		return authorization.Snapshot{}, false, nil
	}

	var (
		status               string
		memberType           string
		authorizationVersion int64
		isOwner              bool
	)
	err := r.pool.QueryRowContext(ctx, `
		SELECT CASE WHEN business.owner_id = ? THEN 'active' ELSE COALESCE(member.status, 'active') END,
		       COALESCE(member.member_type, 'member'),
		       COALESCE(member.authorization_version, 1), business.owner_id = ?
		FROM businesses AS business
		LEFT JOIN authorization_business_members AS member
		  ON member.business_id = business.id AND member.user_id = ?
		WHERE business.id = ? AND (business.owner_id = ? OR member.user_id = ?)
		LIMIT 1`, userID, userID, userID, businessID, userID, userID).
		Scan(&status, &memberType, &authorizationVersion, &isOwner)
	if errors.Is(err, sql.ErrNoRows) {
		return authorization.Snapshot{}, false, nil
	}
	if err != nil {
		return authorization.Snapshot{}, false, err
	}

	roles, err := r.loadAuthorizationRoles(ctx, businessID)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}
	if err := r.loadAuthorizationRolePermissions(ctx, businessID, roles); err != nil {
		return authorization.Snapshot{}, false, err
	}

	isAdmin := memberType == "admin"
	dashPermissions := []string{}

	allPermissions, err := r.loadAuthorizationCatalog(ctx)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}

	licenses, err := r.loadAuthorizationLicenses(ctx, businessID, userID)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}
	assignments, err := r.loadAuthorizationRoleAssignments(ctx, businessID, userID)
	if err != nil {
		return authorization.Snapshot{}, false, err
	}
	for _, assignment := range assignments {
		if assignment.AppID != "dash" || !assignment.Active {
			continue
		}
		for _, profileID := range assignment.RoleIDs {
			for _, permission := range authorization.PermissionsForRole(profileID, roles) {
				dashPermissions = appendUniqueAuthorizationApp(dashPermissions, permission)
			}
		}
	}

	// Owners and admins are tenant administrators and inherit the business's
	// Owners and admins inherit the business plan. Other members need an active
	// app profile assignment and a matching app license.
	inheritsPlan := isOwner || isAdmin
	source := make([]string, 0, len(businessApps))
	if inheritsPlan {
		source = businessApps
	} else {
		for _, assignment := range assignments {
			if assignment.Active {
				source = append(source, string(assignment.AppID))
			}
		}
	}

	apps := make([]authorization.AppID, 0, len(source))
	seenApps := make(map[authorization.AppID]struct{}, len(source))
	for _, rawApp := range source {
		appID := authorization.AppID(rawApp)
		if appID == "" || appID == "dash" {
			continue
		}
		if _, seen := seenApps[appID]; seen {
			continue
		}
		seenApps[appID] = struct{}{}
		apps = append(apps, appID)
	}

	return authorization.Snapshot{
		BusinessID:           businessID,
		IsOwner:              isOwner,
		IsAdmin:              isAdmin,
		MembershipActive:     status == "active",
		SubscribedApps:       apps,
		DashPermissions:      dashPermissions,
		Licenses:             licenses,
		Assignments:          assignments,
		Roles:                roles,
		AllPermissions:       allPermissions,
		AuthorizationVersion: authorizationVersion,
	}, true, nil
}

func (r *Repository) loadAuthorizationRoles(ctx context.Context, businessID int64) ([]authorization.Role, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, app_id, role_key, display_name, description, level,
		       parent_role_id, is_archived, is_system, assignment_mode
		FROM authorization_roles
		WHERE business_id IN (0, ?)
		  AND lifecycle = 'active'
		ORDER BY business_id, app_id, level, id`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]authorization.Role, 0)
	for rows.Next() {
		var (
			role                 authorization.Role
			appID                string
			description          sql.NullString
			parentRoleID         sql.NullInt64
			isArchived, isSystem bool
		)
		if err := rows.Scan(
			&role.ID,
			&appID,
			&role.Key,
			&role.DisplayName,
			&description,
			&role.Level,
			&parentRoleID,
			&isArchived,
			&isSystem,
			&role.AssignmentMode,
		); err != nil {
			return nil, err
		}
		role.AppID = authorizationAppPointer(appID)
		role.ParentRoleID = nullInt64Pointer(parentRoleID)
		role.IsArchived = isArchived
		role.IsSystem = isSystem
		if description.Valid {
			role.Description = description.String
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) loadAuthorizationRolePermissions(ctx context.Context, businessID int64, roles []authorization.Role) error {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT rp.role_id, rp.permission_key
		FROM authorization_role_permissions rp
		JOIN authorization_roles r ON r.id = rp.role_id
		WHERE r.business_id IN (0, ?) AND r.lifecycle = 'active'`, businessID)
	if err != nil {
		return err
	}
	defer rows.Close()

	permissionsByRole := make(map[int64][]string, len(roles))
	for rows.Next() {
		var roleID int64
		var permission string
		if err := rows.Scan(&roleID, &permission); err != nil {
			return err
		}
		permissionsByRole[roleID] = append(permissionsByRole[roleID], permission)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for index := range roles {
		roles[index].Permissions = append([]string{}, permissionsByRole[roles[index].ID]...)
	}
	return nil
}

func (r *Repository) loadAuthorizationCatalog(ctx context.Context) ([]string, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT permission_key
		FROM authorization_permission_catalog
		ORDER BY permission_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]string, 0)
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, rows.Err()
}

func (r *Repository) loadAuthorizationLicenses(ctx context.Context, businessID, userID int64) ([]authorization.License, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT app_id, source, status, group_id, granted_at, revoked_at
		FROM authorization_license_assignments
		WHERE business_id = ?
		  AND status = 'active'
		  AND (user_id = ? OR group_id IN (
			SELECT group_id
			FROM user_group_members
			WHERE business_id = ? AND user_id = ?
		  ))`, businessID, userID, businessID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	licenses := make([]authorization.License, 0)
	for rows.Next() {
		var (
			appID, source, status string
			groupID               sql.NullInt64
			grantedAt, revokedAt  sql.NullTime
		)
		if err := rows.Scan(&appID, &source, &status, &groupID, &grantedAt, &revokedAt); err != nil {
			return nil, err
		}
		license := authorization.License{
			AppID:     authorization.AppID(appID),
			Source:    authorization.AssignmentSource(source),
			Active:    status == "active",
			GroupID:   nullInt64Pointer(groupID),
			GrantedAt: time.Time{}.Format(time.RFC3339),
		}
		if grantedAt.Valid {
			license.GrantedAt = grantedAt.Time.UTC().Format(time.RFC3339)
		}
		if revokedAt.Valid {
			value := revokedAt.Time.UTC().Format(time.RFC3339)
			license.RevokedAt = &value
		}
		licenses = append(licenses, license)
	}
	return licenses, rows.Err()
}
