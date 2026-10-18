package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrCampaignNotSendable covers a campaign that is not email or was
	// already sent or cancelled.
	ErrCampaignNotSendable = errors.New("campaign cannot be sent")
	ErrCampaignIncomplete  = errors.New("campaign needs a subject and a message")
	// ErrSenderAddressRequired: every marketing email must carry the
	// sender's postal address (CAN-SPAM), so a business without one on file
	// cannot send.
	ErrSenderAddressRequired = errors.New("business postal address is required to send marketing email")
	ErrInvalidUnsubscribe    = errors.New("invalid unsubscribe link")
)

// CampaignToSend is the part of a campaign that delivery reads.
type CampaignToSend struct {
	ID        int64
	Name      string
	Subject   string
	Content   string
	SegmentID *int64
}

// AudienceMember is one customer a campaign may reach.
type AudienceMember struct {
	CustomerID int64
	FirstName  string
	LastName   string
	Email      string
	Points     int
	Tier       string
}

// CampaignAudience counts who a campaign reaches: Eligible customers have an
// email address and gave marketing consent, out of Total in the audience.
type CampaignAudience struct {
	Eligible int
	Total    int
}

// CampaignSender is who marketing email comes from.
type CampaignSender struct {
	Name    string
	Address string
	ReplyTo string
}

// UnsubscribeToken signs a business and customer pair so an unsubscribe link
// cannot be forged or pointed at another customer. It never expires: the
// link must keep working for as long as the email exists.
func UnsubscribeToken(secret []byte, businessID, customerID int64) string {
	payload := strconv.FormatInt(businessID, 10) + "." + strconv.FormatInt(customerID, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(unsubscribeMAC(secret, payload))
}

// ParseUnsubscribeToken returns the business and customer a token was signed
// for, or ErrInvalidUnsubscribe.
func ParseUnsubscribeToken(secret []byte, token string) (businessID, customerID int64, err error) {
	encodedPayload, encodedMAC, ok := strings.Cut(token, ".")
	if !ok {
		return 0, 0, ErrInvalidUnsubscribe
	}
	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return 0, 0, ErrInvalidUnsubscribe
	}
	mac, err := base64.RawURLEncoding.DecodeString(encodedMAC)
	if err != nil || !hmac.Equal(mac, unsubscribeMAC(secret, string(payload))) {
		return 0, 0, ErrInvalidUnsubscribe
	}
	business, customer, ok := strings.Cut(string(payload), ".")
	if !ok {
		return 0, 0, ErrInvalidUnsubscribe
	}
	businessID, businessErr := strconv.ParseInt(business, 10, 64)
	customerID, customerErr := strconv.ParseInt(customer, 10, 64)
	if businessErr != nil || customerErr != nil || businessID <= 0 || customerID <= 0 {
		return 0, 0, ErrInvalidUnsubscribe
	}
	return businessID, customerID, nil
}

func unsubscribeMAC(secret []byte, payload string) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = fmt.Fprint(mac, "crm-unsubscribe-v1:", payload)
	return mac.Sum(nil)
}
