package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/subscriptions/internal/application"
	"github.com/voxire/onion/internal/platform/logger"
	paddlePkg "github.com/voxire/onion/internal/platform/paddle"
)

const maxPaddleWebhookBodyBytes = 1 << 20 // 1 MiB

// PaddleWebhookHTTPHandler handles Paddle events with the raw body preserved for HMAC verification.
func PaddleWebhookHTTPHandler(service *application.Service) http.HandlerFunc {
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
			http.Error(w, `{"error":"invalid webhook signature"}`, http.StatusUnauthorized)
			return
		}

		var event paddlePkg.WebhookEvent
		if parseErr := json.Unmarshal(rawBody, &event); parseErr != nil {
			http.Error(w, `{"error":"failed to parse webhook payload"}`, http.StatusBadRequest)
			return
		}

		switch event.EventType {
		case "subscription.created", "subscription.updated", "subscription.activated",
			"subscription.cancelled", "subscription.paused":
		default:
			w.WriteHeader(http.StatusOK)
			return
		}

		var data paddlePkg.SubscriptionData
		if parseErr := json.Unmarshal(event.Data, &data); parseErr != nil {
			http.Error(w, `{"error":"failed to parse subscription data"}`, http.StatusBadRequest)
			return
		}

		if handleErr := service.HandlePaddleWebhook(ctx, event, data); handleErr != nil {
			logger.Ctx(ctx).Error("paddle webhook processing failed",
				logger.String("event_type", event.EventType),
				logger.String("subscription_id", data.ID),
				logger.Error(handleErr),
			)
			http.Error(w, `{"error":"failed to process event"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
