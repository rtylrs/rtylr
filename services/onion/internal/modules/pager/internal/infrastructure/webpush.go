package infrastructure

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

// pushTTL is how long a push service keeps an undelivered message. A phone
// that comes back online later than this has missed the moment anyway.
const pushTTL = 10 * time.Minute

// WebPush sends Web Push messages signed with the server's VAPID key pair.
type WebPush struct {
	publicKey  string
	privateKey string
	subject    string
	client     *http.Client
}

// NewWebPush checks the VAPID key pair: a P-256 public key and private scalar
// in URL-safe base64, as `npx web-push generate-vapid-keys` prints them.
// subject is a mailto: address or https: URL the push services can reach the
// sender at.
func NewWebPush(publicKey, privateKey, subject string) (*WebPush, error) {
	public, err := decodeVAPIDKey(publicKey)
	if err != nil || len(public) != 65 || public[0] != 4 {
		return nil, errors.New("the VAPID public key must be a URL-safe base64 uncompressed P-256 point")
	}
	private, err := decodeVAPIDKey(privateKey)
	if err != nil || len(private) != 32 {
		return nil, errors.New("the VAPID private key must be a URL-safe base64 32 byte P-256 scalar")
	}
	// The library puts mailto: in front of anything that is not an https: URL.
	subject = strings.TrimPrefix(strings.TrimSpace(subject), "mailto:")
	if subject == "" {
		return nil, errors.New("the VAPID subject is empty")
	}
	return &WebPush{
		publicKey:  base64.RawURLEncoding.EncodeToString(public),
		privateKey: base64.RawURLEncoding.EncodeToString(private),
		subject:    subject,
		client: &http.Client{
			Timeout: 10 * time.Second,
			// Endpoints are checked against the known push services when they
			// are stored; a redirect must not lead anywhere else.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

func (w *WebPush) PublicKey() string { return w.publicKey }

// Send delivers one encrypted payload. A push service answering 404 or 410
// has dropped the subscription, reported as domain.ErrPushGone.
func (w *WebPush) Send(ctx context.Context, sub domain.PushSubscription, payload []byte, topic string) error {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{P256dh: sub.P256dh, Auth: sub.Auth},
	}, &webpush.Options{
		HTTPClient:      w.client,
		Subscriber:      w.subject,
		VAPIDPublicKey:  w.publicKey,
		VAPIDPrivateKey: w.privateKey,
		TTL:             int(pushTTL.Seconds()),
		Urgency:         webpush.UrgencyHigh,
		Topic:           topic,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return domain.ErrPushGone
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	default:
		return fmt.Errorf("push service answered %d", resp.StatusCode)
	}
}

func decodeVAPIDKey(key string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(strings.TrimSpace(key), "="))
}
