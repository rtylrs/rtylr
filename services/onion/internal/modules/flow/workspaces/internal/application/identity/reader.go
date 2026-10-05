package identity

import "context"

type IdentityKey struct{ BusinessID, UserID int64 }
type Identity struct{ Name, Email string }

// IdentityReader resolves display identities in a batch. Missing identities are omitted.
// It must propagate read failures and must not be used to authorize a mutation.
type IdentityReader func(context.Context, []IdentityKey) (map[IdentityKey]Identity, error)
