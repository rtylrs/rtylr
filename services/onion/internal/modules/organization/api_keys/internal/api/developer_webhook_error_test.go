package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/platform/hubdeveloper"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func TestWebhookErrorMapsCallerMistakesToClientStatuses(t *testing.T) {
	for name, tc := range map[string]struct {
		err     error
		status  int
		message string
	}{
		"invalid":  {err: fmt.Errorf("%w: url is required", hubdeveloper.ErrInvalidWebhook), status: http.StatusBadRequest, message: "url is required"},
		"missing":  {err: hubdeveloper.ErrWebhookNotFound, status: http.StatusNotFound, message: "Webhook not found"},
		"internal": {err: errors.New("db down"), status: http.StatusInternalServerError, message: "Failed to update webhook"},
	} {
		t.Run(name, func(t *testing.T) {
			var got *onekiterrors.Error
			if !errors.As(webhookError(tc.err, "Failed to update webhook"), &got) {
				t.Fatal("expected a OneKit error")
			}
			if got.HTTPStatusCode() != tc.status || got.GetMessage() != tc.message {
				t.Fatalf("got %d %q, want %d %q", got.HTTPStatusCode(), got.GetMessage(), tc.status, tc.message)
			}
		})
	}
}
