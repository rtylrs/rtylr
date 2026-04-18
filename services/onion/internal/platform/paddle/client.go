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
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/logger"
)

// ── Types ─────────────────────────────────────────────────────────────────────

// CreateTransactionRequest is the body for POST /transactions.
// https://developer.paddle.com/api-reference/transactions/create-transaction
type CreateTransactionRequest struct {
	Items          []TransactionItem    `json:"items"`
	CustomerID     string               `json:"customer_id,omitempty"`
	CustomerEmail  string               `json:"customer_email,omitempty"`
	CheckoutURL    string               `json:"checkout"`
	SuccessURL     string               `json:"success_url,omitempty"`
	CustomData     map[string]string    `json:"custom_data,omitempty"`
	CollectionMode string               `json:"collection_mode"`
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
	EventType        string          `json:"event_type"`
	OccurredAt       time.Time       `json:"occurred_at"`
	Data             json.RawMessage `json:"data"`
}

// SubscriptionData is embedded in subscription.* webhook events.
type SubscriptionData struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	CustomerID string     `json:"customer_id"`
	Items      []struct {
		Price struct {
			ID        string `json:"id"`
			ProductID string `json:"product_id"`
		} `json:"price"`
		Status   string `json:"status"`
		Quantity int    `json:"quantity"`
	} `json:"items"`
	CustomData     map[string]string `json:"custom_data"`
	CurrentBillingPeriod *struct {
		EndsAt time.Time `json:"ends_at"`
	} `json:"current_billing_period"`
}

// ── Client ─────────────────────────────────────────────────────────────────────

// CreateTransaction creates a Paddle transaction (checkout session).
// Returns the checkout URL the user should be redirected to.
func CreateTransaction(ctx context.Context, ownerID int64, priceID string, customerEmail string) (string, error) {
	if !config.IsPaddleConfigured() {
		return "", fmt.Errorf("paddle not configured")
	}

	body := map[string]interface{}{
		"items": []map[string]interface{}{
			{"price_id": priceID, "quantity": 1},
		},
		"collection_mode":  "automatic",
		"customer_email":   customerEmail,
		"custom_data": map[string]string{
			"owner_id": fmt.Sprintf("%d", ownerID),
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
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

// VerifyWebhookSignature validates the Paddle-Signature header against the raw body.
// Returns nil on success, error if the signature is invalid.
//
// Paddle signature format: "ts=<unix_ts>;h1=<hex_hmac>"
// Reference: https://developer.paddle.com/webhooks/signature-verification
func VerifyWebhookSignature(rawBody []byte, signatureHeader string) error {
	secret := config.GetPaddleWebhookSecret()
	if secret == "" {
		// No secret configured — skip verification in development.
		return nil
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
		return fmt.Errorf("invalid Paddle-Signature header format")
	}

	// Compute HMAC-SHA256 of "ts:rawBody"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + ":"))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(h1)) {
		return fmt.Errorf("webhook signature mismatch")
	}

	return nil
}
