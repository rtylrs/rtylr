package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/config/internal/domain"
)

// The update response is merged into client-side config state, so every
// email/SMTP field the PUT accepts has to come back in response.config.
func TestToProtoConfigPayloadPreservesEmailSettings(t *testing.T) {
	payload := newConfigPayload(domain.Config{
		OwnerId:                   "7",
		SenderEmail:               "orders@shop.com",
		EmailProvider:             "smtp",
		SenderName:                "Shop",
		ReplyToEmail:              "support@shop.com",
		EmailNotificationsEnabled: true,
		SMTPHost:                  "smtp.shop.com",
		SMTPPort:                  587,
		SMTPUsername:              "shop",
		SMTPPasswordConfigured:    true,
		BusinessSlug:              "shop",
	})

	got := toProtoConfigPayload(payload)
	if got == nil {
		t.Fatal("toProtoConfigPayload returned nil")
	}
	if got.SenderEmail != "orders@shop.com" {
		t.Errorf("SenderEmail = %q", got.SenderEmail)
	}
	if got.EmailProvider != "smtp" {
		t.Errorf("EmailProvider = %q", got.EmailProvider)
	}
	if got.SenderName != "Shop" {
		t.Errorf("SenderName = %q", got.SenderName)
	}
	if got.ReplyToEmail != "support@shop.com" {
		t.Errorf("ReplyToEmail = %q", got.ReplyToEmail)
	}
	if !got.EmailNotificationsEnabled {
		t.Error("EmailNotificationsEnabled = false")
	}
	if got.SmtpHost != "smtp.shop.com" {
		t.Errorf("SmtpHost = %q", got.SmtpHost)
	}
	if got.SmtpPort != 587 {
		t.Errorf("SmtpPort = %d", got.SmtpPort)
	}
	if got.SmtpUsername != "shop" {
		t.Errorf("SmtpUsername = %q", got.SmtpUsername)
	}
	if !got.SmtpPasswordConfigured {
		t.Error("SmtpPasswordConfigured = false")
	}
}
