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
