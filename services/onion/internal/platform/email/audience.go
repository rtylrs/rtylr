package email

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/resend/resend-go/v2"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrAudienceDisabled is returned when no Resend client is configured, so
// callers can tell "not configured" from a provider failure.
var ErrAudienceDisabled = errors.New("resend audience sync disabled: no RESEND_AUDIENCE_KEY or RESEND_API_KEY")

// ErrAudienceUnauthorized is returned when the configured Resend key is
// rejected (for example, a send-only key returning 401 on Contacts.Create), so
// callers can abort the run without burning per-row retry attempts.
var ErrAudienceUnauthorized = errors.New("resend audience sync unauthorized")

// IsAudienceAuthError reports whether err is a Resend configuration,
// authentication, or permission failure (such as a missing key or a send-only
// key returning 401 Unauthorized / "This API key is restricted to only send
// emails").
func IsAudienceAuthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrAudienceDisabled) || errors.Is(err, ErrAudienceUnauthorized) {
		return true
	}
	msg := strings.ToLower(err.Error())
	for _, token := range []string{
		"401",
		"403",
		"unauthorized",
		"forbidden",
		"restricted to only send emails",
		"restricted_api_key",
		"invalid api key",
		"api key is invalid",
	} {
		if strings.Contains(msg, token) {
			return true
		}
	}
	return false
}

// audienceClient talks to the account that owns the newsletter audience.
// That is not always the sending account: voxire.com's transactional mail
// goes out on one Resend account while the blog audience lives on another,
// so RESEND_AUDIENCE_KEY, when set, gets its own client. Unset, the sending
// client is reused.
var audienceClient *resend.Client

func initAudienceClient() {
	var key string
	if config.GlobalConfig != nil {
		key = config.GlobalConfig.ResendAudienceKey
	} else {
		key = os.Getenv("RESEND_AUDIENCE_KEY")
	}
	if key == "" {
		audienceClient = client
		return
	}
	audienceClient = resend.NewClient(key)
	logger.Global().Info("Resend audience client initialized")
}

// AddAudienceContact mirrors one address into a Resend audience and returns
// the contact id. Resend treats a repeated email inside one audience as the
// same contact, so calling this twice for one address is safe. The key must
// have contact permissions: a "sending access" key gets a 401 here, which is
// wrapped with ErrAudienceUnauthorized so the caller can abort the run without
// burning per-row retry attempts.
func AddAudienceContact(ctx context.Context, audienceID, address string) (string, error) {
	if audienceClient == nil {
		return "", ErrAudienceDisabled
	}
	if err := waitForOutboundSlot(ctx); err != nil {
		return "", err
	}
	res, err := audienceClient.Contacts.CreateWithContext(ctx, &resend.CreateContactRequest{
		Email:      strings.ToLower(strings.TrimSpace(address)),
		AudienceId: audienceID,
	})
	if err != nil {
		if IsAudienceAuthError(err) {
			return "", fmt.Errorf("%w: %w", ErrAudienceUnauthorized, err)
		}
		return "", err
	}
	return res.Id, nil
}
