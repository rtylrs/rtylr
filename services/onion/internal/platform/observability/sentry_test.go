package observability

import (
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestScrubRequestDataRemovesBodyAndQuery(t *testing.T) {
	event := &sentry.Event{Request: &sentry.Request{
		Data:        `{"password":"secret"}`,
		QueryString: "token=secret",
		URL:         "https://example.test/invite?token=secret",
		Cookies:     "session=secret",
	}}

	if got := scrubRequestData(event, nil); got != event {
		t.Fatal("scrubber should preserve the event")
	}
	if event.Request.Data != "" || event.Request.QueryString != "" || event.Request.Cookies != "" || event.Request.URL != "https://example.test/invite" {
		t.Fatalf("request data was not scrubbed: %+v", event.Request)
	}
}

func TestScrubRequestDataHandlesMissingRequest(t *testing.T) {
	event := &sentry.Event{}
	if got := scrubRequestData(event, nil); got != event {
		t.Fatal("scrubber should preserve events without request metadata")
	}
}
