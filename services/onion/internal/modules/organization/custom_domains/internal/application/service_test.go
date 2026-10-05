package application

import (
	"context"
	"errors"
	"testing"
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
