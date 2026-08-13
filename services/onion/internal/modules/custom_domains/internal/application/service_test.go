package application

import "testing"

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateDomain(tt.value); (err != nil) != tt.wantErr {
				t.Fatalf("validateDomain(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
