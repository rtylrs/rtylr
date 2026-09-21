package infrastructure

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
)

// ── Manual Addon Grants CRUD ─────────────────────────────────────────────────

func (r *Repository) ListAddons(ctx context.Context) ([]AddonCatalogRow, error) {
	query, args, err := sq.
		Select("id", "name", "description", "icon", "price", "is_free", "sort_order").
		From("addons").
		OrderBy("sort_order ASC", "id ASC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addons := make([]AddonCatalogRow, 0)
	for rows.Next() {
		var a AddonCatalogRow
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon, &a.Price, &a.IsFree, &a.SortOrder); err != nil {
			return nil, err
		}
		addons = append(addons, a)
	}
	return addons, rows.Err()
}

func (r *Repository) AddonExists(ctx context.Context, addonID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*)").
		From("addons").
		Where(sq.Eq{"id": addonID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// dynamic-query: squirrel
func (r *Repository) ListAddonGrants(ctx context.Context, limit, offset int) ([]AddonGrantRow, int64, error) {
	countSQL, countArgs, err := sq.
		Select("COUNT(*)").
		From("manual_addon_grants").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL, listArgs, err := sq.
		Select(
			"g.id", "g.user_id", "COALESCE(u.full_name, '')", "u.email",
			"g.business_id", "COALESCE(b.name, '')",
			"g.addon_id", "a.name", "a.price",
			"g.granted_by", "adm.full_name", "g.note", "g.expires_at", "g.revoked_at",
			"g.enabled_by_grant", "g.created_at",
		).
		From("manual_addon_grants g").
		Join("addons a ON a.id = g.addon_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.id = g.business_id").
		LeftJoin("users adm ON adm.id = g.granted_by").
		OrderBy("g.created_at DESC").
		Limit(uint64(limit)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	grants := make([]AddonGrantRow, 0, limit)
	for rows.Next() {
		var g AddonGrantRow
		if err := rows.Scan(&g.ID, &g.UserID, &g.UserName, &g.UserEmail, &g.BusinessID, &g.BusinessName,
			&g.AddonID, &g.AddonName, &g.AddonPrice, &g.GrantedBy, &g.GranterName, &g.Note,
			&g.ExpiresAt, &g.RevokedAt, &g.EnabledByGrant, &g.CreatedAt); err != nil {
			return nil, 0, err
		}
		grants = append(grants, g)
	}
	return grants, total, rows.Err()
}

func (r *Repository) CreateAddonGrant(ctx context.Context, userID, businessID int64, addonID string, grantedBy int64, note string, expiresAt *time.Time, enabledByGrant bool) (int64, error) {
	var expiresSQL interface{}
	if expiresAt != nil {
		expiresSQL = *expiresAt
	}
	query, args, err := sq.
		Insert("manual_addon_grants").
		Columns("user_id", "business_id", "addon_id", "granted_by", "note", "expires_at", "enabled_by_grant").
		Values(userID, businessID, addonID, grantedBy, note, expiresSQL, enabledByGrant).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	res, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAddonGrant returns the display fields for one grant without loading the
// full grant history. CreateAddonGrant uses it to build its response.
func (r *Repository) GetAddonGrant(ctx context.Context, id int64) (*AddonGrantRow, error) {
	query, args, err := sq.
		Select(
			"g.id", "g.user_id", "COALESCE(u.full_name, '')", "u.email",
			"g.business_id", "COALESCE(b.name, '')",
			"g.addon_id", "a.name", "a.price",
			"g.granted_by", "adm.full_name", "g.note", "g.expires_at", "g.revoked_at",
			"g.enabled_by_grant", "g.created_at",
		).
		From("manual_addon_grants g").
		Join("addons a ON a.id = g.addon_id").
		LeftJoin("users u ON u.id = g.user_id").
		LeftJoin("businesses b ON b.id = g.business_id").
		LeftJoin("users adm ON adm.id = g.granted_by").
		Where(sq.Eq{"g.id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	var grant AddonGrantRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(
		&grant.ID, &grant.UserID, &grant.UserName, &grant.UserEmail, &grant.BusinessID, &grant.BusinessName,
		&grant.AddonID, &grant.AddonName, &grant.AddonPrice, &grant.GrantedBy, &grant.GranterName, &grant.Note,
		&grant.ExpiresAt, &grant.RevokedAt, &grant.EnabledByGrant, &grant.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &grant, nil
}

func (r *Repository) GetAddonGrantScope(ctx context.Context, id int64) (*AddonGrantScopeRow, error) {
	query, args, err := sq.
		Select("user_id", "business_id", "addon_id", "enabled_by_grant").
		From("manual_addon_grants").
		Where(sq.Eq{"id": id}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	var scope AddonGrantScopeRow
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&scope.UserID, &scope.BusinessID, &scope.AddonID, &scope.EnabledByGrant); err != nil {
		return nil, err
	}
	return &scope, nil
}

// dynamic-query: squirrel
func (r *Repository) RevokeAddonGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Update("manual_addon_grants").
		Set("revoked_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) DeleteAddonGrant(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("manual_addon_grants").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) IsBusinessAddonEnabled(ctx context.Context, businessID int64, addonID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*)").
		From("business_addons").
		Where(sq.Eq{"business_id": businessID, "addon_id": addonID, "is_enabled": true}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) EnableBusinessAddon(ctx context.Context, businessID int64, addonID string) error {
	query := `
INSERT INTO business_addons (business_id, addon_id, is_enabled, enabled_at)
VALUES (?, ?, true, NOW())
ON DUPLICATE KEY UPDATE is_enabled = true, enabled_at = NOW(), disabled_at = NULL`
	_, err := r.pool.ExecContext(ctx, query, businessID, addonID)
	return err
}

// dynamic-query: squirrel
func (r *Repository) DisableBusinessAddon(ctx context.Context, businessID int64, addonID string) error {
	query, args, err := sq.
		Update("business_addons").
		Set("is_enabled", false).
		Set("disabled_at", sq.Expr("NOW()")).
		Where(sq.Eq{"business_id": businessID, "addon_id": addonID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) HasActiveAddonGrant(ctx context.Context, businessID int64, addonID string) (bool, error) {
	query, args, err := sq.
		Select("COUNT(*)").
		From("manual_addon_grants").
		Where(sq.Eq{"business_id": businessID, "addon_id": addonID, "revoked_at": nil}).
		Where("(expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return false, err
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// AdoptAddonGrantEnablement moves the "this grant turned the add-on on" marker
// to a surviving active grant. Revoking keeps the marker row around, but
// deleting hard-removes it, and either way a later removal of the remaining
// grant must still be able to tell a grant-driven enablement from one the
// business enabled itself.
func (r *Repository) AdoptAddonGrantEnablement(ctx context.Context, businessID int64, addonID string, removedID int64) error {
	query := `
UPDATE manual_addon_grants
SET enabled_by_grant = TRUE
WHERE business_id = ? AND addon_id = ? AND id <> ?
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
  AND enabled_by_grant = FALSE
ORDER BY id ASC
LIMIT 1`
	_, err := r.pool.ExecContext(ctx, query, businessID, addonID, removedID)
	return err
}
