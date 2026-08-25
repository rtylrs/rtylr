package application

import (
	"errors"
	"testing"

	"github.com/workos/workos-go/v4/pkg/workos_errors"
)

func TestIsWorkOSUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "transport", err: errors.New("connection reset"), want: true},
		{name: "rate limited", err: workos_errors.HTTPError{Code: 429}, want: true},
		{name: "provider failure", err: workos_errors.HTTPError{Code: 503}, want: true},
		{name: "invalid code", err: workos_errors.HTTPError{Code: 400}, want: false},
		{name: "expired code", err: workos_errors.HTTPError{Code: 401}, want: false},
		{name: "structured client error", err: &workos_errors.SSORequiredError{HTTPError: workos_errors.HTTPError{Code: 403}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWorkOSUnavailable(tt.err); got != tt.want {
				t.Fatalf("isWorkOSUnavailable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
