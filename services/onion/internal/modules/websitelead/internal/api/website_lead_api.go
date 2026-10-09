// Package api serves the public website's lead capture over the
// website.lead.v1 OneKit contract.
package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/voxire/onion/internal/modules/websitelead/internal/application"
	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	websiteleadv1 "github.com/voxire/voxie/gen/go/website/lead/v1"
)

// WebsiteLeadAPI implements websiteleadv1.WebsiteLeadServiceServer. Required
// fields and lengths are enforced by the generated Validate(); the rules here
// are the ones the schema cannot express.
type WebsiteLeadAPI struct {
	Service *application.Service
}

var _ websiteleadv1.WebsiteLeadServiceServer = (*WebsiteLeadAPI)(nil)

func (a *WebsiteLeadAPI) SubmitQuote(ctx context.Context, req *websiteleadv1.SubmitQuoteRequest) (*websiteleadv1.SubmitQuoteResponse, error) {
	// Prepend the timeline to the brief so it surfaces in the notification email
	// and Google Sheets without a dedicated column.
	project := req.GetProject()
	if req.GetTimeline() != "" {
		project = fmt.Sprintf("Timeline: %s\n\n%s", req.GetTimeline(), strings.TrimSpace(req.GetProject()))
	}
	id, err := a.Service.Submit(ctx, domain.LeadInput{
		Name:    req.GetName(),
		Email:   req.GetEmail(),
		Phone:   req.GetPhoneNumber(),
		Company: req.GetCompany(),
		Service: req.GetService(),
		Project: project,
		Budget:  req.GetBudget(),
		Source:  req.GetSource(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to submit quote request")
	}
	return &websiteleadv1.SubmitQuoteResponse{Id: id, Message: "Quote request received"}, nil
}

func (a *WebsiteLeadAPI) Contact(ctx context.Context, req *websiteleadv1.ContactRequest) (*websiteleadv1.ContactResponse, error) {
	_, err := a.Service.Submit(ctx, domain.LeadInput{
		Name:    req.GetName(),
		Email:   req.GetEmail(),
		Phone:   req.GetPhoneNumber(),
		Service: "contact",
		Project: req.GetMessage(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to submit contact request")
	}
	return &websiteleadv1.ContactResponse{Message: "Message received"}, nil
}

func (a *WebsiteLeadAPI) SubscribeNewsletter(ctx context.Context, req *websiteleadv1.NewsletterRequest) (*websiteleadv1.NewsletterResponse, error) {
	if err := normalizeNewsletter(req); err != nil {
		return nil, err
	}
	id, err := a.Service.Subscribe(ctx, domain.SubscriberInput{
		Email:        req.GetEmail(),
		Source:       req.GetSource(),
		PostSlug:     req.GetPostSlug(),
		PostTitle:    req.GetPostTitle(),
		PostCategory: req.GetPostCategory(),
		Language:     req.GetLanguage(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to subscribe")
	}
	return &websiteleadv1.NewsletterResponse{Id: id, Message: "Subscribed"}, nil
}

// RecordAnalyticsEvent logs whatever arrives and always succeeds, so the
// website's fire-and-forget call never surfaces an error.
func (a *WebsiteLeadAPI) RecordAnalyticsEvent(ctx context.Context, req *websiteleadv1.AnalyticsEventRequest) (*websiteleadv1.AnalyticsEventResponse, error) {
	logger.Ctx(ctx).Info("website analytics event",
		logger.Any("event", req.GetEvent()),
		logger.Any("page", req.GetPage()),
		logger.Any("job_id", string(req.GetJobId())),
		logger.Any("job_title", req.GetJobTitle()),
	)
	return &websiteleadv1.AnalyticsEventResponse{Logged: true}, nil
}

// simpleEmail is deliberately loose: one @, something on both sides, a dot in
// the domain. Deliverability is Resend's problem, not the form's.
var simpleEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// normalizeNewsletter trims and lower-cases the email and defaults the
// language in place before checking, so the stored row is what was validated.
func normalizeNewsletter(req *websiteleadv1.NewsletterRequest) error {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Source = strings.TrimSpace(req.Source)
	req.Language = strings.TrimSpace(req.Language)
	if req.Language == "" {
		req.Language = "en"
	}
	bad := func(message string) error { return onekiterrors.New(http.StatusBadRequest, message) }
	switch {
	case req.Email == "":
		return bad("email is required")
	case len(req.Email) > 255 || !simpleEmail.MatchString(req.Email):
		return bad("email is invalid")
	case req.Source == "":
		return bad("source is required")
	case !domain.IsAllowedNewsletterSource(req.Source):
		return bad("source is invalid")
	case !domain.IsAllowedLanguage(req.Language):
		return bad("language is invalid")
	}
	return nil
}
