package email

import (
	"context"
	"errors"
	"strings"

	"github.com/resend/resend-go/v2"
)

// ErrAudienceDisabled is returned when no Resend client is configured, so
// callers can tell "not configured" from a provider failure.
var ErrAudienceDisabled = errors.New("resend audience sync disabled: no RESEND_API_KEY")

// AddAudienceContact mirrors one address into a Resend audience and returns
// the contact id. Resend treats a repeated email inside one audience as the
// same contact, so calling this twice for one address is safe. The key must
// have contact permissions: a "sending access" key gets a 401 here, which the
// caller records on the row rather than retrying forever.
func AddAudienceContact(ctx context.Context, audienceID, address string) (string, error) {
	if client == nil {
		return "", ErrAudienceDisabled
	}
	if err := waitForOutboundSlot(ctx); err != nil {
		return "", err
	}
	res, err := client.Contacts.CreateWithContext(ctx, &resend.CreateContactRequest{
		Email:      strings.ToLower(strings.TrimSpace(address)),
		AudienceId: audienceID,
	})
	if err != nil {
		return "", err
	}
	return res.Id, nil
}
