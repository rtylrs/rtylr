package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
)

func TestSubscribe_Delegates(t *testing.T) {
	// A successful subscribe fires one detached notification. Wait for it so
	// no goroutine outlives the seam swap.
	_, wait := captureEmails(t, 1)
	m := &mockRepo{subscribeID: 3}
	s := &Service{repo: m, notify: testNotify()}
	id, err := s.Subscribe(context.Background(), domain.SubscriberInput{Email: "r@example.com", Source: "intel_drop", Language: "en"})
	if err != nil || id != 3 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if m.subscribeGot.Email != "r@example.com" {
		t.Fatalf("repo got %+v", m.subscribeGot)
	}
	wait()
}

func TestSubscribe_PropagatesError(t *testing.T) {
	captureEmails(t, 0)
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{subscribeErr: want}, notify: testNotify()}
	if _, err := s.Subscribe(context.Background(), domain.SubscriberInput{Email: "r@example.com"}); !errors.Is(err, want) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}

func TestSubscribe_NotifiesNewsletterListOnly(t *testing.T) {
	got, wait := captureEmails(t, 1)
	s := &Service{repo: &mockRepo{subscribeID: 11}, notify: testNotify()}
	_, err := s.Subscribe(context.Background(), domain.SubscriberInput{
		Email: "rania@example.com", Source: "article_pdf_mid", PostSlug: "seo-lebanon-2026",
		PostTitle: "SEO in Lebanon", PostCategory: "SEO", Language: "ar",
	})
	if err != nil {
		t.Fatal(err)
	}
	wait()
	if len(*got) != 1 {
		t.Fatalf("expected exactly one email, got %d", len(*got))
	}
	p := (*got)[0]
	if strings.Join(p.To, ",") != "news@example.com" {
		t.Fatalf("To = %v, want the newsletter list", p.To)
	}
	if p.Subject != "[Newsletter] rania@example.com via article_pdf_mid" {
		t.Fatalf("Subject = %q", p.Subject)
	}
	if p.From != "Voxire Website <noreply@voxire.com>" || p.ReplyTo != "rania@example.com" {
		t.Fatalf("From=%q ReplyTo=%q", p.From, p.ReplyTo)
	}
	for _, want := range []string{"SEO in Lebanon", "seo-lebanon-2026", "SEO", "ar", "#11"} {
		if !strings.Contains(p.HTML, want) {
			t.Fatalf("HTML missing %q", want)
		}
	}
	if strings.Contains(p.HTML, "\u2014") {
		t.Fatal("HTML contains an em dash")
	}
}

func TestSubscribe_EncodesSlugInLink(t *testing.T) {
	got, wait := captureEmails(t, 1)
	s := &Service{repo: &mockRepo{subscribeID: 21}, notify: testNotify()}
	_, err := s.Subscribe(context.Background(), domain.SubscriberInput{
		Email: "rania@example.com", Source: "article_pdf_mid", PostSlug: "seo lebanon?x=1",
	})
	if err != nil {
		t.Fatal(err)
	}
	wait()
	p := (*got)[0]
	if !strings.Contains(p.HTML, "https://voxire.com/blog/seo%20lebanon%3Fx=1/") {
		t.Fatalf("HTML missing encoded slug link: %s", p.HTML)
	}
	if strings.Contains(p.HTML, "blog/seo lebanon") {
		t.Fatal("HTML contains an unencoded slug in the link")
	}
}
