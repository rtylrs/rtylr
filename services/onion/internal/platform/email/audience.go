package email

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/resend/resend-go/v2"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrAudienceDisabled is returned when no Resend client is configured, so
// callers can tell "not configured" from a provider failure.
var ErrAudienceDisabled = errors.New("resend audience sync disabled: no RESEND_AUDIENCE_KEY or RESEND_API_KEY")

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
// have contact permissions: a "sending access" key gets a 401 here, which the
// caller records on the row rather than retrying forever.
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
		return "", err
	}
	return res.Id, nil
}
