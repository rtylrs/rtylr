// Package paddle provides a minimal client for the Paddle Billing API.
// Docs: https://developer.paddle.com/api-reference
package paddle

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// ── Types ─────────────────────────────────────────────────────────────────────

// CreateTransactionRequest is the body for POST /transactions.
// https://developer.paddle.com/api-reference/transactions/create-transaction
type CreateTransactionRequest struct {
	Items          []TransactionItem `json:"items"`
	CustomerID     string            `json:"customer_id,omitempty"`
	CustomerEmail  string            `json:"customer_email,omitempty"`
	CheckoutURL    string            `json:"checkout"`
	SuccessURL     string            `json:"success_url,omitempty"`
	CustomData     map[string]string `json:"custom_data,omitempty"`
	CollectionMode string            `json:"collection_mode"`
}

// TransactionItem is a single line in a Paddle transaction.
type TransactionItem struct {
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

// TransactionResponse is the relevant subset of the Paddle transaction response.
type TransactionResponse struct {
	Data struct {
		ID       string `json:"id"`
		Checkout struct {
			URL string `json:"url"`
		} `json:"checkout"`
		Status string `json:"status"`
	} `json:"data"`
	Error *struct {
		Type   string `json:"type"`
		Code   string `json:"code"`
		Detail string `json:"detail"`
	} `json:"error"`
}

// WebhookEvent is the outer envelope of all Paddle webhook notifications.
type WebhookEvent struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

// SubscriptionData is embedded in subscription.* webhook events.
type SubscriptionData struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	CustomerID string `json:"customer_id"`
	Items      []struct {
		Price struct {
			ID        string `json:"id"`
			ProductID string `json:"product_id"`
		} `json:"price"`
		Status   string `json:"status"`
		Quantity int    `json:"quantity"`
	} `json:"items"`
	CustomData           map[string]string `json:"custom_data"`
	CurrentBillingPeriod *struct {
		EndsAt time.Time `json:"ends_at"`
	} `json:"current_billing_period"`
}

// ── Client ─────────────────────────────────────────────────────────────────────

// POSPaymentRequest represents a POS-originated Paddle transaction creation.
type POSPaymentRequest struct {
	OwnerID        int64
	OrderID        int64
	AmountMinor    int64  // amount in minor units (e.g. cents)
	Currency       string // ISO currency code (USD, EUR, etc.)
	CustomerEmail  string
	Description    string
	IdempotencyKey string
}

// POSPaymentResult is returned by CreatePOSTransaction.
type POSPaymentResult struct {
	TransactionID string
	CheckoutURL   string
}

const (
	paddleRequestTimeout       = 15 * time.Second
	maxPaddleResponseBodyBytes = 1 << 20
)

var paddleHTTPClient = &http.Client{Timeout: paddleRequestTimeout}

// CreatePOSTransaction creates a one-off Paddle transaction for a POS order.
// Uses the idempotency key so retries with the same key resolve to the same
// transaction on Paddle's side. The key is mandatory: without it a
// transport-level retry would silently create a duplicate charge.
func CreatePOSTransaction(ctx context.Context, req POSPaymentRequest) (POSPaymentResult, error) {
	if !config.IsPaddleConfigured() {
		return POSPaymentResult{}, errors.New("paddle not configured")
	}
	if req.IdempotencyKey == "" {
		return POSPaymentResult{}, errors.New("paddle idempotency key is required")
	}

	// Paddle's transaction-creation endpoint accepts an inline one-off price via
	// the items array with non_catalog_price payloads.
	body := map[string]interface{}{
		"collection_mode": "automatic",
		"customer_email":  req.CustomerEmail,
		"custom_data": map[string]string{
			"owner_id": strconv.FormatInt(req.OwnerID, 10),
			"order_id": strconv.FormatInt(req.OrderID, 10),
		},
		"items": []map[string]interface{}{
			{
				"quantity": 1,
				"price": map[string]interface{}{
					"description": req.Description,
					"unit_price": map[string]interface{}{
						"amount":        strconv.FormatInt(req.AmountMinor, 10),
						"currency_code": strings.ToUpper(req.Currency),
					},
				},
			},
		},
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return POSPaymentResult{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		config.GetPaddleBaseURL()+"/transactions", bytes.NewReader(raw))
	if err != nil {
		return POSPaymentResult{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.GetPaddleAPIKey())
	httpReq.Header.Set("Paddle-Idempotency-Key", req.IdempotencyKey)

	resp, err := paddleHTTPClient.Do(httpReq)
	if err != nil {
		return POSPaymentResult{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := readPaddleResponseBody(resp.Body)
	if err != nil {
		return POSPaymentResult{}, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		logger.Ctx(ctx).Error("paddle create POS transaction failed",
			logger.String("status", resp.Status),
			logger.String("body", string(respBody)),
		)
		return POSPaymentResult{}, fmt.Errorf("paddle API returned %s", resp.Status)
	}

	var txResp TransactionResponse
	if err := json.Unmarshal(respBody, &txResp); err != nil {
		return POSPaymentResult{}, err
	}
	if txResp.Error != nil {
		return POSPaymentResult{}, fmt.Errorf("paddle error %s: %s", txResp.Error.Code, txResp.Error.Detail)
	}

	return POSPaymentResult{
		TransactionID: txResp.Data.ID,
		CheckoutURL:   txResp.Data.Checkout.URL,
	}, nil
}

// CreateTransaction creates a Paddle transaction (checkout session).
// Returns the checkout URL the user should be redirected to.
func CreateTransaction(ctx context.Context, businessID int64, priceID, customerEmail string) (string, error) {
	if !config.IsPaddleConfigured() {
		return "", errors.New("paddle not configured")
	}

	body := map[string]interface{}{
		"items": []map[string]interface{}{
			{"price_id": priceID, "quantity": 1},
		},
		"collection_mode": "automatic",
		"customer_email":  customerEmail,
		"custom_data": map[string]string{
			"business_id": strconv.FormatInt(businessID, 10),
		},
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		config.GetPaddleBaseURL()+"/transactions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.GetPaddleAPIKey())

	resp, err := paddleHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := readPaddleResponseBody(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		logger.Ctx(ctx).Error("paddle create transaction failed",
			logger.String("status", resp.Status),
			logger.String("body", string(respBody)),
		)

		return "", fmt.Errorf("paddle API returned %s", resp.Status)
	}

	var txResp TransactionResponse
	if err := json.Unmarshal(respBody, &txResp); err != nil {
		return "", err
	}

	if txResp.Error != nil {
		return "", fmt.Errorf("paddle error %s: %s", txResp.Error.Code, txResp.Error.Detail)
	}

	return txResp.Data.Checkout.URL, nil
}

func readPaddleResponseBody(body io.Reader) ([]byte, error) {
	respBody, err := io.ReadAll(io.LimitReader(body, maxPaddleResponseBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(respBody) > maxPaddleResponseBodyBytes {
		return nil, errors.New("paddle response body too large")
	}
	return respBody, nil
}

const webhookSignatureTolerance = 5 * time.Minute

// VerifyWebhookSignature validates the Paddle-Signature header against the raw body.
// Returns nil on success, error if the signature is invalid.
//
// Paddle signature format: "ts=<unix_ts>;h1=<hex_hmac>"
// Reference: https://developer.paddle.com/webhooks/signature-verification
func VerifyWebhookSignature(rawBody []byte, signatureHeader string) error {
	secret := config.GetPaddleWebhookSecret()
	if secret == "" {
		// Fail closed: both webhook routes are anonymous, so an empty secret
		// would accept forged billing payloads. Set PADDLE_WEBHOOK_SECRET, or
		// opt out explicitly for local work with PADDLE_ALLOW_INSECURE_WEBHOOKS
		// (refused outside development).
		if os.Getenv("PADDLE_ALLOW_INSECURE_WEBHOOKS") == "1" && !config.IsProduction() {
			return nil
		}
		return errors.New("PADDLE_WEBHOOK_SECRET is required to verify webhook signatures")
	}

	var ts, h1 string
	for _, part := range strings.Split(signatureHeader, ";") {
		if strings.HasPrefix(part, "ts=") {
			ts = strings.TrimPrefix(part, "ts=")
		}
		if strings.HasPrefix(part, "h1=") {
			h1 = strings.TrimPrefix(part, "h1=")
		}
	}

	if ts == "" || h1 == "" {
		return errors.New("invalid Paddle-Signature header format")
	}

	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return errors.New("invalid Paddle-Signature timestamp")
	}
	signedAt := time.Unix(timestamp, 0)
	if delta := time.Since(signedAt); delta > webhookSignatureTolerance || delta < -webhookSignatureTolerance {
		return errors.New("Paddle-Signature timestamp outside tolerance")
	}

	// Compute HMAC-SHA256 of "ts:rawBody"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + ":"))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(h1)) {
		return errors.New("webhook signature mismatch")
	}

	return nil
}
