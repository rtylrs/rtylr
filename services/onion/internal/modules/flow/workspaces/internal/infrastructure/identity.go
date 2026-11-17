package infrastructure

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
)

func (r *Repository) WithIdentityReaders(reader identity.IdentityReader, email func(context.Context, string) (int64, error), username func(context.Context, int64, string) (int64, error)) *Repository {
	r.identities, r.accountByEmail, r.businessUserByUsername = reader, email, username
	return r
}

func (r *Repository) readIdentities(ctx context.Context, keys []identity.IdentityKey) (map[identity.IdentityKey]identity.Identity, error) {
	if len(keys) == 0 {
		return map[identity.IdentityKey]identity.Identity{}, nil
	}
	if r.identities == nil {
		return nil, errors.New("Flow identity reader is not configured")
	}
	return r.identities(ctx, keys)
}

func displayName(p identity.Identity) string {
	if p.Name != "" {
		return p.Name
	}
	return p.Email
}

// AssigneeNames returns the display name Flow's task list shows for each of
// userIDs in businessID: the identity's name, read through the same identity
// reader that sets a listed task's AssigneeName. Users with no identity or no
// name are left out.
func (r *Repository) AssigneeNames(ctx context.Context, businessID int64, userIDs []int64) (map[int64]string, error) {
	keys := make([]identity.IdentityKey, 0, len(userIDs))
	for _, id := range userIDs {
		if businessID > 0 && id > 0 {
			keys = append(keys, identity.IdentityKey{BusinessID: businessID, UserID: id})
		}
	}
	profiles, err := r.readIdentities(ctx, keys)
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(profiles))
	for key, profile := range profiles {
		if key.BusinessID == businessID && profile.Name != "" {
			names[key.UserID] = profile.Name
		}
	}
	return names, nil
}
