package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/domain"
)

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid subdomain", value: "menu.example.com"},
		{name: "empty", value: "", wantErr: true},
		{name: "path", value: "example.com/menu", wantErr: true},
		{name: "port", value: "example.com:443", wantErr: true},
		{name: "missing suffix", value: "localhost", wantErr: true},
		{name: "wildcard", value: "*.example.com", wantErr: true},
		{name: "underscore", value: "_acme.example.com", wantErr: true},
		{name: "leading hyphen", value: "-menu.example.com", wantErr: true},
		{name: "trailing hyphen", value: "menu-.example.com", wantErr: true},
		{name: "platform domain", value: "api.voxire.dev", wantErr: true},
		{name: "platform subdomain", value: "tenant.rtylr.com", wantErr: true},
		{name: "platform apex", value: "tenant.voxire.com", wantErr: true},
		{name: "ip address", value: "192.0.2.1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateDomain(tt.value); (err != nil) != tt.wantErr {
				t.Fatalf("validateDomain(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

type fakeTXTResolver struct {
	records map[string][]string
	err     error
}

func (r fakeTXTResolver) LookupTXT(_ context.Context, name string) ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.records[name], nil
}

func TestDNSChallengeMatches(t *testing.T) {
	const (
		domainName = "menu.example.com"
		token      = "challenge-token"
	)
	name := "_rtylr-challenge." + domainName

	tests := []struct {
		name     string
		resolver txtResolver
		want     bool
	}{
		{name: "matching record", resolver: fakeTXTResolver{records: map[string][]string{name: {"other", token}}}, want: true},
		{name: "mismatched record", resolver: fakeTXTResolver{records: map[string][]string{name: {"other"}}}},
		{name: "lookup failure", resolver: fakeTXTResolver{err: errors.New("dns unavailable")}},
		{name: "nil resolver", resolver: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{resolver: tt.resolver}
			if got := service.dnsChallengeMatches(context.Background(), domainName, token); got != tt.want {
				t.Fatalf("dnsChallengeMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}

type stubRepo struct {
	current    domain.CustomDomain
	registered bool
}

func (r *stubRepo) Get(context.Context, int64) (domain.CustomDomain, error) {
	if r.current.Domain == "" {
		return domain.CustomDomain{}, domain.ErrNotFound
	}
	return r.current, nil
}

func (r *stubRepo) GetByDomain(_ context.Context, name string) (domain.CustomDomain, error) {
	if r.current.Domain == name {
		return r.current, nil
	}
	return domain.CustomDomain{}, domain.ErrNotFound
}

func (r *stubRepo) Register(_ context.Context, input domain.CustomDomain) (domain.CustomDomain, error) {
	r.registered = true
	input.Status = "pending_verification"
	return input, nil
}

func (r *stubRepo) Remove(context.Context, int64) error         { return nil }
func (r *stubRepo) Verify(context.Context, int64, string) error { return nil }

func TestRegisterKeepsVerifiedDomainVerified(t *testing.T) {
	repo := &stubRepo{current: domain.CustomDomain{BusinessID: 1, Domain: "menu.example.com", Status: domain.StatusVerified, VerificationToken: "t"}}
	got, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "Menu.Example.com.")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repo.registered {
		t.Fatal("re-registering a verified domain must not rotate its challenge")
	}
	if got.Status != domain.StatusVerified {
		t.Fatalf("status = %q, want verified", got.Status)
	}
}

func TestRegisterRotatesChallengeForPendingDomain(t *testing.T) {
	repo := &stubRepo{current: domain.CustomDomain{BusinessID: 1, Domain: "menu.example.com", Status: "pending_verification", VerificationToken: "old"}}
	got, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "menu.example.com")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if !repo.registered || got.VerificationToken == "old" {
		t.Fatal("a pending domain should get a fresh challenge")
	}
}
