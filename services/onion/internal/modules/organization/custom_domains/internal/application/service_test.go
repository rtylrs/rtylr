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
	claims     []domain.CustomDomain
	registered *domain.CustomDomain
	verified   bool
	verifyErr  error
}

func (r *stubRepo) Get(_ context.Context, businessID int64) (domain.CustomDomain, error) {
	for _, claim := range r.claims {
		if claim.BusinessID == businessID {
			return claim, nil
		}
	}
	return domain.CustomDomain{}, domain.ErrNotFound
}

func (r *stubRepo) GetVerifiedByDomain(_ context.Context, name string) (domain.CustomDomain, error) {
	for _, claim := range r.claims {
		if claim.Domain == name && claim.Status == domain.StatusVerified {
			return claim, nil
		}
	}
	return domain.CustomDomain{}, domain.ErrNotFound
}

func (r *stubRepo) Register(_ context.Context, input domain.CustomDomain) (domain.CustomDomain, error) {
	r.registered = &input
	input.Status = "pending_verification"
	return input, nil
}

func (r *stubRepo) Remove(context.Context, int64) error { return nil }
func (r *stubRepo) Verify(context.Context, int64, string) error {
	r.verified = r.verifyErr == nil
	return r.verifyErr
}

func TestRegisterKeepsVerifiedDomainVerified(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 1, Domain: "menu.example.com", Status: domain.StatusVerified, VerificationToken: "t"}}}
	got, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "Menu.Example.com.")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repo.registered != nil {
		t.Fatal("re-registering a verified domain must not rotate its challenge")
	}
	if got.Status != domain.StatusVerified {
		t.Fatalf("status = %q, want verified", got.Status)
	}
}

func TestRegisterRotatesChallengeForPendingDomain(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 1, Domain: "menu.example.com", Status: "pending_verification", VerificationToken: "old"}}}
	got, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "menu.example.com")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repo.registered == nil || got.VerificationToken == "old" {
		t.Fatal("a pending domain should get a fresh challenge")
	}
}

func TestRegisterIsNotBlockedByAnotherBusinessUnverifiedClaim(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 2, Domain: "menu.example.com", Status: "pending_verification", VerificationToken: "squatter"}}}
	if _, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "menu.example.com"); err != nil {
		t.Fatalf("an unverified claim blocked the real owner: %v", err)
	}
	if repo.registered == nil || repo.registered.BusinessID != 1 {
		t.Fatalf("registered = %+v", repo.registered)
	}
}

func TestRegisterIsBlockedByAnotherBusinessVerifiedClaim(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 2, Domain: "menu.example.com", Status: domain.StatusVerified}}}
	if _, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "menu.example.com"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if repo.registered != nil {
		t.Fatal("a verified domain was claimed by another business")
	}
}

func TestRegisterAnotherDomainReplacesThePendingClaim(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 1, Domain: "old.example.com", Status: domain.StatusVerified}}}
	if _, err := NewService(context.Background(), repo).Register(context.Background(), 9, 1, "new.example.com"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repo.registered == nil || repo.registered.Domain != "new.example.com" {
		t.Fatalf("registered = %+v", repo.registered)
	}
}

func verifiableService(repo *stubRepo, records map[string][]string) *Service {
	service := NewService(context.Background(), repo)
	service.resolver = fakeTXTResolver{records: records}
	return service
}

func TestVerifyRequiresTheDNSChallenge(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 1, Domain: "menu.example.com", Status: "pending_verification", VerificationToken: "tok"}}}
	if _, err := verifiableService(repo, nil).Verify(context.Background(), 1, "tok"); !errors.Is(err, domain.ErrDNSVerificationFailed) {
		t.Fatalf("err = %v, want ErrDNSVerificationFailed", err)
	}
	if repo.verified {
		t.Fatal("verified without a DNS record")
	}
}

func TestVerifyRejectsAWrongToken(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 1, Domain: "menu.example.com", VerificationToken: "tok"}}}
	records := map[string][]string{"_rtylr-challenge.menu.example.com": {"tok"}}
	if _, err := verifiableService(repo, records).Verify(context.Background(), 1, "other"); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyAcceptsAQuotedTXTRecord(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{{BusinessID: 1, Domain: "menu.example.com", VerificationToken: "tok"}}}
	records := map[string][]string{"_rtylr-challenge.menu.example.com": {`"tok"`}}
	if _, err := verifiableService(repo, records).Verify(context.Background(), 1, "tok"); err != nil || !repo.verified {
		t.Fatalf("err = %v verified = %v", err, repo.verified)
	}
}

func TestVerifyLosesToAnotherBusinessThatVerifiedFirst(t *testing.T) {
	repo := &stubRepo{claims: []domain.CustomDomain{
		{BusinessID: 2, Domain: "menu.example.com", Status: domain.StatusVerified, VerificationToken: "theirs"},
		{BusinessID: 1, Domain: "menu.example.com", Status: "pending_verification", VerificationToken: "tok"},
	}}
	records := map[string][]string{"_rtylr-challenge.menu.example.com": {"tok"}}
	if _, err := verifiableService(repo, records).Verify(context.Background(), 1, "tok"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if repo.verified {
		t.Fatal("a second business verified an already verified domain")
	}
}
