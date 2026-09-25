package api

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func NewsletterRoutes() []string { return []string{"/newsletter"} }

type NewsletterHandler struct{ Service *application.Service }

type NewsletterRequest struct {
	Email        string `json:"email"`
	Source       string `json:"source"`
	PostSlug     string `json:"post_slug"`
	PostTitle    string `json:"post_title"`
	PostCategory string `json:"post_category"`
	Language     string `json:"language"`
}

type NewsletterResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// simpleEmail is deliberately loose: one @, something on both sides, a dot in
// the domain. Deliverability is Resend's problem, not the form's.
var simpleEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Validate normalises the request in place (trimmed lower-case email, default
// language) before checking it, so Serve stores exactly what was validated.
func (h *NewsletterHandler) Validate(_ context.Context, req *NewsletterRequest) error {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Source = strings.TrimSpace(req.Source)
	req.Language = strings.TrimSpace(req.Language)
	if req.Language == "" {
		req.Language = "en"
	}

	if req.Email == "" {
		return onekiterrors.New(http.StatusBadRequest, "email is required")
	}
	if len(req.Email) > 255 || !simpleEmail.MatchString(req.Email) {
		return onekiterrors.New(http.StatusBadRequest, "email is invalid")
	}
	if req.Source == "" {
		return onekiterrors.New(http.StatusBadRequest, "source is required")
	}
	if !domain.IsAllowedNewsletterSource(req.Source) {
		return onekiterrors.New(http.StatusBadRequest, "source is invalid")
	}
	if !domain.IsAllowedLanguage(req.Language) {
		return onekiterrors.New(http.StatusBadRequest, "language is invalid")
	}
	if len(req.PostSlug) > 255 {
		return onekiterrors.New(http.StatusBadRequest, "post_slug is too long")
	}
	if len(req.PostTitle) > 255 {
		return onekiterrors.New(http.StatusBadRequest, "post_title is too long")
	}
	if len(req.PostCategory) > 100 {
		return onekiterrors.New(http.StatusBadRequest, "post_category is too long")
	}
	return nil
}

func (h *NewsletterHandler) Serve(ctx context.Context, req *NewsletterRequest) (NewsletterResponse, error) {
	id, err := h.Service.Subscribe(ctx, domain.SubscriberInput{
		Email:        req.Email,
		Source:       req.Source,
		PostSlug:     req.PostSlug,
		PostTitle:    req.PostTitle,
		PostCategory: req.PostCategory,
		Language:     req.Language,
	})
	if err != nil {
		return NewsletterResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to subscribe")
	}
	return NewsletterResponse{ID: id, Message: "Subscribed"}, nil
}
