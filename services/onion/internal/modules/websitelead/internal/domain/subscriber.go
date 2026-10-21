package domain

// SubscriberInput is one newsletter signup from voxire.com.
type SubscriberInput struct {
	Email        string
	Source       string
	PostSlug     string
	PostTitle    string
	PostCategory string
	Language     string
}

// PendingSubscriber is a row still missing from the Resend audience.
type PendingSubscriber struct {
	ID    int64
	Email string
}

// AllowedNewsletterSources are the form placements on voxire.com that can
// produce a signup. Anything else is rejected at validation.
var AllowedNewsletterSources = []string{
	"intel_drop",
	"report_banner",
	"article_pdf_mid",
	"article_pdf_end",
}

func IsAllowedNewsletterSource(s string) bool {
	for _, allowed := range AllowedNewsletterSources {
		if s == allowed {
			return true
		}
	}
	return false
}

// IsAllowedLanguage accepts the two languages the site is served in.
func IsAllowedLanguage(s string) bool {
	return s == "en" || s == "ar"
}
