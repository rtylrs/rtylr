package identityadapter

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/organization/scopes"
)

func NewIdentityReader(reader *scopes.Reader) identity.IdentityReader {
	return func(ctx context.Context, keys []identity.IdentityKey) (map[identity.IdentityKey]identity.Identity, error) {
		grouped := make(map[int64]map[int64]struct{})
		for _, key := range keys {
			if key.BusinessID <= 0 || key.UserID <= 0 {
				continue
			}
			if grouped[key.BusinessID] == nil {
				grouped[key.BusinessID] = make(map[int64]struct{})
			}
			grouped[key.BusinessID][key.UserID] = struct{}{}
		}
		result := make(map[identity.IdentityKey]identity.Identity, len(keys))
		for businessID, ids := range grouped {
			requested := make([]int64, 0, len(ids))
			for id := range ids {
				requested = append(requested, id)
			}
			profiles, err := reader.BusinessUsers(ctx, businessID, requested)
			if err != nil {
				return nil, err
			}
			for _, p := range profiles {
				result[identity.IdentityKey{BusinessID: businessID, UserID: p.ID}] = identity.Identity{Name: p.Name, Email: p.Email}
			}
		}
		return result, nil
	}
}
