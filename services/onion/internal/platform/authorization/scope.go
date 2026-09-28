package authorization

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
)

// ErrInvalidScope is returned for a scope the assignment tables cannot store.
// It matches the chk_auth_assignment_scope_type constraint, which permits only
// pos_location and flow_project with a positive numeric id.
var ErrInvalidScope = errors.New("invalid assignment scope")

// ScopeQueryer is the subset of *sql.Tx the scope validator needs. Validation
// runs inside the caller's transaction so a rejected scope rolls nothing back
// and surfaces as a 4xx rather than a 500 from the constraint.
type ScopeQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ValidateAssignmentScope checks a scope before it is written to
// authorization_role_assignment_scopes. It mirrors the table's CHECK constraint
// and additionally confirms the referenced row belongs to the same business, so
// a scope cannot point at another tenant's location or project.
//
// Every writer of that table must call this. The constraint is the last
// backstop, and it fires late: the normalized team writer performs the license
// and role revocation in the same transaction, so an unvalidated scope would
// roll back a member's whole authorization and return an opaque 500.
func ValidateAssignmentScope(ctx context.Context, tx ScopeQueryer, businessID int64, scopeType, scopeID string) error {
	id, err := strconv.ParseInt(scopeID, 10, 64)
	if err != nil || id <= 0 {
		return ErrInvalidScope
	}
	var query string
	switch scopeType {
	case "pos_location":
		query = `SELECT 1 FROM pos_locations WHERE business_id = ? AND id = ? LIMIT 1`
	case "flow_project":
		query = `SELECT 1 FROM flow_projects WHERE business_id = ? AND id = ? LIMIT 1`
	default:
		return ErrInvalidScope
	}
	if err := tx.QueryRowContext(ctx, query, businessID, id).Scan(new(int)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidScope
		}
		return err
	}
	return nil
}
