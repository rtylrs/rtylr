package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/webhooks/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/hubdeveloper"
)

type mockRepo struct {
	listResult []domain.Webhook
	listErr    error
	// Delivery listing, added with the dead-letter surface. deliveriesGot
	// records the filter so a test can assert the service passes it through
	// unchanged rather than quietly dropping a caller's narrowing.
	deliveriesResult []domain.DeliveryRecord
	deliveriesTotal  int64
	deliveriesErr    error
	deliveriesGot    struct {
		businessID int64
		filter     domain.DeliveryFilter
	}
	createID  int64
	createErr error
	getResult domain.Webhook
	getErr    error
	updateErr error
	deleteErr error
	rotateErr error
	rotateGot struct {
		businessID int64
		id         int64
		secret     string
	}
	createGot struct {
		ownerID    int64
		businessID int64
		url        string
		secret     string
		events     []string
	}
	deleteGot struct {
		businessID int64
		id         int64
	}
	updateGot struct {
		businessID int64
		id         int64
		url        string
		events     []string
		isActive   bool
	}
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Webhook, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) Create(_ context.Context, ownerID, businessID int64, url, secret string, events []string) (int64, error) {
	m.createGot.ownerID = ownerID
	m.createGot.businessID = businessID
	m.createGot.url = url
	m.createGot.secret = secret
	m.createGot.events = events
	return m.createID, m.createErr
}

func (m *mockRepo) Get(_ context.Context, _, _ int64) (domain.Webhook, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) Update(_ context.Context, businessID, id int64, url string, events []string, isActive bool) error {
	m.updateGot.businessID = businessID
	m.updateGot.id = id
	m.updateGot.url = url
	m.updateGot.events = events
	m.updateGot.isActive = isActive
	return m.updateErr
}

func (m *mockRepo) RotateSecret(_ context.Context, businessID, id int64, secret string) error {
	m.rotateGot.businessID = businessID
	m.rotateGot.id = id
	m.rotateGot.secret = secret
	return m.rotateErr
}

func (m *mockRepo) Delete(_ context.Context, businessID, id int64) error {
	m.deleteGot.businessID = businessID
	m.deleteGot.id = id
	return m.deleteErr
}

func TestList_Delegates(t *testing.T) {
	want := []domain.Webhook{{ID: 1, URL: "https://x"}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, err := s.List(context.Background(), 1)
	if err != nil || len(got) != 1 || got[0].URL != "https://x" {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestCreate_HappyPathPopulatesSecretAndEvents(t *testing.T) {
	m := &mockRepo{createID: 50}
	s := &Service{repo: m}
	id, secret, err := s.Create(context.Background(), 7, 7, "https://hooks.example.com/in", []string{"order_created"})
	if err != nil || id != 50 {
		t.Fatalf("err=%v id=%d", err, id)
	}
	if secret != m.createGot.secret {
		t.Errorf("returned secret %q is not the stored secret %q", secret, m.createGot.secret)
	}
	if m.createGot.url != "https://hooks.example.com/in" || m.createGot.ownerID != 7 {
		t.Errorf("repo got %+v", m.createGot)
	}
	if len(m.createGot.secret) != 64 {
		t.Errorf("expected 64-char hex secret, got %d", len(m.createGot.secret))
	}
	if len(m.createGot.events) != 1 || m.createGot.events[0] != "order_created" {
		t.Errorf("events: %v", m.createGot.events)
	}
}

func TestCreate_NilEventsNormalizedToEmpty(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	if _, _, err := s.Create(context.Background(), 1, 1, "https://x.example.com", nil); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.createGot.events == nil || len(m.createGot.events) != 0 {
		t.Errorf("expected empty events slice, got %v", m.createGot.events)
	}
}

func TestCreate_PropagatesRepoError(t *testing.T) {
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, _, err := s.Create(context.Background(), 1, 1, "https://x.example.com", nil); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_DelegatesIDs(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if err := s.Delete(context.Background(), 7, 42); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if m.deleteGot.businessID != 7 || m.deleteGot.id != 42 {
		t.Errorf("delete args: %+v", m.deleteGot)
	}
}

func TestGenerateSecret_HexAndUnique(t *testing.T) {
	a, err := generateSecret()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, _ := generateSecret()
	if a == b {
		t.Fatal("two secrets should differ")
	}
	if len(a) != 64 {
		t.Errorf("expected 64-char hex secret, got %d", len(a))
	}
}

func TestCreate_RequiresURL(t *testing.T) {
	s := &Service{}
	if _, _, err := s.Create(context.Background(), 1, 1, "", nil); err == nil {
		t.Fatalf("expected error for empty url")
	}
}

func TestDelete_RequiresWebhookID(t *testing.T) {
	s := &Service{}
	if err := s.Delete(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing webhook id")
	}
}

func (m *mockRepo) ListDeliveriesForBusiness(
	_ context.Context, businessID int64, filter domain.DeliveryFilter,
) ([]domain.DeliveryRecord, int64, error) {
	m.deliveriesGot.businessID = businessID
	m.deliveriesGot.filter = filter
	return m.deliveriesResult, m.deliveriesTotal, m.deliveriesErr
}

type staticEventReader struct{ event eventbus.Event }

func (r staticEventReader) GetEvent(context.Context, uint64, int64) (eventbus.Event, error) {
	return r.event, nil
}

// The service is a pass-through here, and that is the point worth pinning: the
// business scope and the caller's filter must both reach the repository, since
// the repository is where the tenant scope is enforced.
func TestListDeliveriesPassesTheScopeAndFilterThrough(t *testing.T) {
	m := &mockRepo{
		deliveriesResult: []domain.DeliveryRecord{{ID: 7, EventID: 9, Status: "dead"}},
		deliveriesTotal:  1,
	}
	service := &Service{repo: m, events: staticEventReader{event: eventbus.Event{ID: 9, Type: "order_created"}}}

	filter := domain.DeliveryFilter{WebhookID: 42, Status: "dead", Limit: 25, Offset: 50}
	got, total, err := service.ListDeliveries(context.Background(), 9003, filter)
	if err != nil {
		t.Fatalf("ListDeliveries: %v", err)
	}
	if total != 1 || len(got) != 1 || got[0].ID != 7 || got[0].EventType != "order_created" {
		t.Errorf("got %+v (total %d), want the single record with its event type", got, total)
	}
	if m.deliveriesGot.businessID != 9003 {
		t.Errorf("business = %d, want 9003: the tenant scope did not reach the repository", m.deliveriesGot.businessID)
	}
	if m.deliveriesGot.filter != filter {
		t.Errorf("filter = %+v, want %+v", m.deliveriesGot.filter, filter)
	}
}

func TestListDeliveriesPropagatesAFailure(t *testing.T) {
	m := &mockRepo{deliveriesErr: errListDeliveries}
	service := &Service{repo: m}

	if _, _, err := service.ListDeliveries(context.Background(), 9003, domain.DeliveryFilter{}); err == nil {
		t.Fatal("a repository failure was swallowed")
	}
}

var errListDeliveries = errors.New("listing failed")

func TestRegistrationValidation(t *testing.T) {
	for name, tc := range map[string]struct {
		url    string
		events []string
		ok     bool
	}{
		"https":             {url: "https://hooks.example.com/in", events: []string{"order_created"}, ok: true},
		"http":              {url: "http://hooks.example.com", ok: true},
		"padded":            {url: "  https://hooks.example.com  ", ok: true},
		"empty":             {url: "  "},
		"javascript":        {url: "javascript:alert(1)"},
		"ftp":               {url: "ftp://hooks.example.com"},
		"no host":           {url: "https:///path"},
		"credentials":       {url: "https://user:pw@hooks.example.com"},
		"localhost":         {url: "http://localhost:8080/hook"},
		"loopback":          {url: "http://127.0.0.1/hook"},
		"metadata":          {url: "http://169.254.169.254/latest/meta-data"},
		"private":           {url: "http://10.0.0.5/hook"},
		"ipv6 loopback":     {url: "http://[::1]/hook"},
		"too long":          {url: "https://hooks.example.com/" + strings.Repeat("a", 2048)},
		"unsupported event": {url: "https://hooks.example.com", events: []string{"order.paid"}},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := validateRegistration(tc.url, tc.events)
			if (err == nil) != tc.ok {
				t.Fatalf("validateRegistration(%q) err = %v, ok want %v", tc.url, err, tc.ok)
			}
			if err != nil && !errors.Is(err, hubdeveloper.ErrInvalidWebhook) {
				t.Fatalf("error %v does not wrap ErrInvalidWebhook", err)
			}
		})
	}
}

func TestRegistrationTrimsURLAndDedupesEvents(t *testing.T) {
	url, events, err := validateRegistration(" https://hooks.example.com ", []string{"order_created", " order_created"})
	if err != nil || url != "https://hooks.example.com" || len(events) != 1 {
		t.Fatalf("got %q %v %v", url, events, err)
	}
}

func TestCreateRejectsInvalidRegistrationBeforeStoring(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	if _, _, err := s.Create(context.Background(), 1, 1, "http://169.254.169.254/", nil); !errors.Is(err, hubdeveloper.ErrInvalidWebhook) {
		t.Fatalf("err = %v, want ErrInvalidWebhook", err)
	}
	if m.createGot.url != "" {
		t.Fatal("an invalid registration reached the repository")
	}
}

func TestUpdateRejectsInvalidRegistrationBeforeStoring(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if _, err := s.Update(context.Background(), 1, 5, "ftp://x.example.com", nil, true); !errors.Is(err, hubdeveloper.ErrInvalidWebhook) {
		t.Fatalf("err = %v, want ErrInvalidWebhook", err)
	}
	if m.updateGot.id != 0 {
		t.Fatal("an invalid update reached the repository")
	}
}

func TestRotateSecretStoresAndReturnsTheNewSecret(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	secret, err := s.RotateSecret(context.Background(), 7, 12)
	if err != nil {
		t.Fatalf("RotateSecret: %v", err)
	}
	if len(secret) != 64 || m.rotateGot.secret != secret || m.rotateGot.id != 12 || m.rotateGot.businessID != 7 {
		t.Fatalf("secret=%q stored=%+v", secret, m.rotateGot)
	}
	other, _ := s.RotateSecret(context.Background(), 7, 12)
	if other == secret {
		t.Fatal("rotation returned the same secret twice")
	}
}

func TestRotateSecretReportsAMissingWebhook(t *testing.T) {
	s := &Service{repo: &mockRepo{rotateErr: domain.ErrWebhookNotFound}}
	if _, err := s.RotateSecret(context.Background(), 7, 12); !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("err = %v, want ErrWebhookNotFound", err)
	}
}

func TestRotateSecretRequiresAnID(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	if _, err := s.RotateSecret(context.Background(), 7, 0); err == nil {
		t.Fatal("expected an error for a missing id")
	}
	if m.rotateGot.secret != "" {
		t.Fatal("a secret was written without an id")
	}
}
