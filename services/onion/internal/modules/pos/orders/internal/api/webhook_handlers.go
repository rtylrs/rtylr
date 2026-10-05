package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
	"github.com/voxire/onion/internal/platform/logger"
	paddlePkg "github.com/voxire/onion/internal/platform/paddle"
)

const maxPaddleWebhookBodyBytes = 1 << 20 // 1 MiB

// PaddlePOSWebhookHTTPHandler returns a raw http.HandlerFunc for Paddle webhook
// events targeting POS orders. It bypasses the typed handler pipeline so the
// raw request body is available for HMAC-SHA256 signature verification before
// any JSON parsing occurs.
func PaddlePOSWebhookHTTPHandler(service *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		r.Body = http.MaxBytesReader(w, r.Body, maxPaddleWebhookBodyBytes)
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, `{"error":"request body too large"}`, http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, `{"error":"failed to read request body"}`, http.StatusBadRequest)
			return
		}

		sig := r.Header.Get("Paddle-Signature")
		if verifyErr := paddlePkg.VerifyWebhookSignature(rawBody, sig); verifyErr != nil {
			logger.Ctx(ctx).Warn("rejected paddle POS webhook with invalid signature",
				logger.Error(verifyErr),
			)
			http.Error(w, `{"error":"invalid webhook signature"}`, http.StatusUnauthorized)
			return
		}

		var event paddlePkg.WebhookEvent
		if parseErr := json.Unmarshal(rawBody, &event); parseErr != nil {
			http.Error(w, `{"error":"failed to parse webhook payload"}`, http.StatusBadRequest)
			return
		}

		if event.EventType != "transaction.completed" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Extract transaction ID from the event data.
		var tx struct {
			ID string `json:"id"`
		}
		if parseErr := json.Unmarshal(event.Data, &tx); parseErr != nil {
			http.Error(w, `{"error":"failed to parse transaction data"}`, http.StatusBadRequest)
			return
		}

		if handleErr := service.HandlePaddleTransactionCompleted(ctx, tx.ID, event.EventType); handleErr != nil {
			logger.Ctx(ctx).Error("paddle POS webhook processing failed",
				logger.String("event_type", event.EventType),
				logger.String("transaction_id", tx.ID),
				logger.Error(handleErr),
			)
			// Unexpected errors must 500 so Paddle retries; acking 200 here
			// permanently dropped billing events on transient failures (#755).
			http.Error(w, `{"error":"failed to process transaction"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
