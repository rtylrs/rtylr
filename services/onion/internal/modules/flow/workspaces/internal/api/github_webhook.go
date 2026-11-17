package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	maxGitHubWebhookBodyBytes = 1 << 20 // 1 MiB
	// X-GitHub-Delivery is a GUID and X-GitHub-Event a short name; both are
	// stored, so anything longer than the columns is refused.
	maxGitHubWebhookHeaderLen = 64
)

// GitHubWebhookHTTPHandler receives GitHub App webhooks. It is anonymous: the
// X-Hub-Signature-256 HMAC over the raw body, keyed by the app's webhook
// secret, is the only authentication, so the body is read here rather than by
// a JSON decoder. Without a secret every delivery is refused.
func GitHubWebhookHTTPHandler(service *application.Service, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if len(secret) == 0 {
			logger.Ctx(ctx).Warn("github webhook refused: GITHUB_APP_WEBHOOK_SECRET is not set")
			writeGitHubWebhook(w, http.StatusServiceUnavailable, `{"error":"github webhooks are not configured"}`)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxGitHubWebhookBodyBytes)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeGitHubWebhook(w, http.StatusRequestEntityTooLarge, `{"error":"request body too large"}`)
				return
			}
			writeGitHubWebhook(w, http.StatusBadRequest, `{"error":"failed to read request body"}`)
			return
		}
		if !verifyGitHubWebhookSignature(secret, body, r.Header.Get("X-Hub-Signature-256")) {
			writeGitHubWebhook(w, http.StatusUnauthorized, `{"error":"invalid webhook signature"}`)
			return
		}
		deliveryID := r.Header.Get("X-GitHub-Delivery")
		event := r.Header.Get("X-GitHub-Event")
		if deliveryID == "" || event == "" || len(deliveryID) > maxGitHubWebhookHeaderLen || len(event) > maxGitHubWebhookHeaderLen {
			writeGitHubWebhook(w, http.StatusBadRequest, `{"error":"missing or invalid delivery headers"}`)
			return
		}
		outcome, err := service.HandleGitHubWebhook(ctx, deliveryID, event, body)
		switch {
		case errors.Is(err, domain.ErrGitHubWebhookPayload):
			writeGitHubWebhook(w, http.StatusBadRequest, `{"error":"failed to parse webhook payload"}`)
		case err != nil:
			// Not recorded, so GitHub's redelivery can retry it.
			logger.Ctx(ctx).Error("github webhook processing failed",
				logger.String("event", event),
				logger.String("delivery", deliveryID),
				logger.Error(err),
			)
			writeGitHubWebhook(w, http.StatusInternalServerError, `{"error":"failed to process event"}`)
		case outcome == domain.GitHubWebhookIgnored:
			writeGitHubWebhook(w, http.StatusAccepted, `{"status":"ignored"}`)
		case outcome == domain.GitHubWebhookDuplicate:
			writeGitHubWebhook(w, http.StatusOK, `{"status":"duplicate"}`)
		default:
			writeGitHubWebhook(w, http.StatusOK, `{"status":"processed"}`)
		}
	}
}

func writeGitHubWebhook(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

// verifyGitHubWebhookSignature checks X-Hub-Signature-256, which GitHub sets
// to "sha256=" and the hex HMAC-SHA256 of the raw body keyed by the webhook
// secret. The comparison is constant time. An empty secret never verifies.
func verifyGitHubWebhookSignature(secret, body []byte, header string) bool {
	if len(secret) == 0 {
		return false
	}
	digest, ok := strings.CutPrefix(header, "sha256=")
	if !ok {
		return false
	}
	got, err := hex.DecodeString(digest)
	if err != nil || len(got) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}
