package tenant

import (
	"context"
	"errors"
	"testing"
)

type scopeStub struct {
	ownerID, businessID int64
	err                 error
}

func (s scopeStub) ActiveBusinessOwner(context.Context, int64) (int64, error) {
	return s.ownerID, s.err
}
func (s scopeStub) UserBusiness(context.Context, int64, int64) (int64, error) {
	return s.businessID, s.err
}

func TestResolveOwnerIDFailsClosed(t *testing.T) {
	for _, test := range []struct {
		name       string
		reader     ScopeResolver
		businessID int64
		want       error
	}{
		{"missing scope", scopeStub{ownerID: 7}, 0, ErrNoBusinessScope},
		{"negative scope", scopeStub{ownerID: 7}, -1, ErrNoBusinessScope},
		{"missing resolver", nil, 42, ErrOwnerLookupFailed},
		{"failed lookup", scopeStub{ownerID: 7, err: errors.New("unavailable")}, 42, ErrOwnerLookupFailed},
		{"invalid owner", scopeStub{}, 42, ErrOwnerLookupFailed},
	} {
		t.Run(test.name, func(t *testing.T) {
			id, err := resolveOwnerID(context.Background(), test.reader, test.businessID, 4242)
			if id != 0 || !errors.Is(err, test.want) {
				t.Fatalf("owner=%d err=%v, want zero and %v", id, err, test.want)
			}
		})
	}
	id, err := resolveOwnerID(context.Background(), scopeStub{ownerID: 7}, 42, 4242)
	if err != nil || id != 7 {
		t.Fatalf("verified owner=%d, err=%v", id, err)
	}
}

func TestResolveBusinessIDFailsClosed(t *testing.T) {
	for _, test := range []struct {
		name   string
		reader ScopeResolver
		userID int64
	}{
		{"missing resolver", nil, 7}, {"unauthenticated", scopeStub{businessID: 42}, 0},
		{"failed lookup", scopeStub{businessID: 42, err: errors.New("unavailable")}, 7},
		{"invalid result", scopeStub{}, 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			id, err := resolveBusinessID(context.Background(), test.reader, 42, test.userID)
			if id != 0 || !errors.Is(err, ErrNoBusinessScope) {
				t.Fatalf("business=%d err=%v", id, err)
			}
		})
	}
	restore := SetScopeResolver(scopeStub{businessID: 42, ownerID: 7})
	defer restore()
	id, err := ResolveBusinessID(context.Background(), 42, 7)
	if err != nil || id != 42 {
		t.Fatalf("verified business=%d err=%v", id, err)
	}
}
