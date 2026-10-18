package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

// mockRepo implements the repository interface for unit testing.
type mockRepo struct {
	leads       []domain.Lead
	lead        domain.Lead
	pipeline    []domain.PipelineSummary
	contacts    []domain.Contact
	contact     domain.Contact
	touchpoints []domain.Touchpoint
	touchpoint  domain.Touchpoint
	contracts   []domain.Contract
	contract    domain.Contract
	err         error
}

type fixedBusinessOwner struct{ ownerID int64 }

func (f fixedBusinessOwner) GetBusinessOwnerID(context.Context, int64) (int64, error) {
	return f.ownerID, nil
}

func (m *mockRepo) ListLeads(_ context.Context, _ int64, _, _ int32) ([]domain.Lead, error) {
	return m.leads, m.err
}

func (m *mockRepo) CountLeads(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.leads)), m.err
}

func (m *mockRepo) GetLead(_ context.Context, _, _ int64) (domain.Lead, error) {
	return m.lead, m.err
}

func (m *mockRepo) CreateLead(_ context.Context, _ int64, _ domain.CreateLeadRequest) (domain.Lead, error) {
	return m.lead, m.err
}

func (m *mockRepo) UpdateLead(_ context.Context, _, _ int64, _ domain.UpdateLeadRequest) (domain.Lead, error) {
	return m.lead, m.err
}

func (m *mockRepo) ValidateLeadAssignee(_ context.Context, _, _ int64) error {
	return m.err
}

func (m *mockRepo) ListReps(_ context.Context, _ int64) ([]domain.Rep, error) {
	return []domain.Rep{{ID: 1, Name: "Dana Owner"}}, m.err
}

func (m *mockRepo) DeleteLead(_ context.Context, _, _ int64) error {
	return m.err
}

func (m *mockRepo) GetPipeline(_ context.Context, _ int64) ([]domain.PipelineSummary, error) {
	return m.pipeline, m.err
}

func (m *mockRepo) ListContacts(_ context.Context, _ int64, _ *int64, _, _ int32) ([]domain.Contact, error) {
	return m.contacts, m.err
}

func (m *mockRepo) CountContacts(_ context.Context, _ int64, _ *int64) (int64, error) {
	return int64(len(m.contacts)), m.err
}

func (m *mockRepo) CreateContact(_ context.Context, _, _ int64, _ domain.CreateContactRequest) (domain.Contact, error) {
	return m.contact, m.err
}

func (m *mockRepo) UpdateContact(_ context.Context, _, _ int64, _ domain.UpdateContactRequest) (domain.Contact, error) {
	return m.contact, m.err
}

func (m *mockRepo) DeleteContact(_ context.Context, _, _ int64) error {
	return m.err
}

func (m *mockRepo) ListTouchpoints(_ context.Context, _, _ int64) ([]domain.Touchpoint, error) {
	return m.touchpoints, m.err
}

func (m *mockRepo) CreateTouchpoint(_ context.Context, _, _ int64, _ domain.CreateTouchpointRequest) (domain.Touchpoint, error) {
	return m.touchpoint, m.err
}

func (m *mockRepo) DeleteTouchpoint(_ context.Context, _, _ int64) error {
	return m.err
}

func (m *mockRepo) ListContracts(_ context.Context, _ int64, _, _ int32) ([]domain.Contract, error) {
	return m.contracts, m.err
}

func (m *mockRepo) CountContracts(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.contracts)), m.err
}

func (m *mockRepo) GetContract(_ context.Context, _, _ int64) (domain.Contract, error) {
	return m.contract, m.err
}

func (m *mockRepo) CreateContract(_ context.Context, _ int64, _ domain.CreateContractRequest) (domain.Contract, error) {
	return m.contract, m.err
}

func (m *mockRepo) UpdateContract(_ context.Context, _, _ int64, _ domain.UpdateContractRequest) (domain.Contract, error) {
	return m.contract, m.err
}

// --- Lead tests ---

func TestListLeads_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{leads: []domain.Lead{{ID: 1, CompanyName: "Acme"}}}}
	got, err := s.ListLeads(context.Background(), 10, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Leads) != 1 || got.Leads[0].ID != 1 {
		t.Errorf("expected 1 lead, got %v", got.Leads)
	}
	if got.Total != 1 || got.Limit != 20 || got.Offset != 0 {
		t.Errorf("unexpected pagination metadata: %+v", got)
	}
}

func TestListLeads_PropagatesError(t *testing.T) {
	want := errors.New("db down")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.ListLeads(context.Background(), 10, 0, 0)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestGetLead_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{lead: domain.Lead{ID: 5, CompanyName: "Beta"}}}
	got, err := s.GetLead(context.Background(), 5, 10)
	if err != nil || got.ID != 5 {
		t.Errorf("got %+v, err %v", got, err)
	}
}

func TestGetLead_PropagatesError(t *testing.T) {
	want := domain.ErrLeadNotFound
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetLead(context.Background(), 99, 10)
	if !errors.Is(err, want) {
		t.Errorf("expected ErrLeadNotFound, got %v", err)
	}
}

func TestCreateLead_AppliesDefaults(t *testing.T) {
	m := &mockRepo{lead: domain.Lead{ID: 1, CompanyName: "Unknown", Source: "cold_email", Stage: domain.LeadStageCold}}
	s := &Service{repo: m}
	got, err := s.CreateLead(context.Background(), 10, domain.CreateLeadRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CompanyName != "Unknown" {
		t.Errorf("expected default company name, got %q", got.CompanyName)
	}
}

func TestCreateLead_PropagatesError(t *testing.T) {
	want := errors.New("create failed")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateLead(context.Background(), 10, domain.CreateLeadRequest{CompanyName: "X"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateLead_AppliesDefaultsAndPropagates(t *testing.T) {
	m := &mockRepo{lead: domain.Lead{ID: 3, Stage: domain.LeadStageCold}}
	s := &Service{repo: m}
	_, err := s.UpdateLead(context.Background(), 3, 10, domain.UpdateLeadRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteLead_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteLead(context.Background(), 1, 10); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteLead_PropagatesError(t *testing.T) {
	want := errors.New("delete error")
	s := &Service{repo: &mockRepo{err: want}}
	if err := s.DeleteLead(context.Background(), 1, 10); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetPipeline_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{pipeline: []domain.PipelineSummary{{Stage: "cold", Count: 3}}}}
	got, err := s.GetPipeline(context.Background(), 1)
	if err != nil || len(got) != 1 {
		t.Errorf("got %v, err %v", got, err)
	}
}

// --- Contact tests ---

func TestListContacts_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{contacts: []domain.Contact{{ID: 7}}}}
	got, err := s.ListContacts(context.Background(), 1, nil, 0, 0)
	if err != nil || len(got.Contacts) != 1 {
		t.Errorf("got %v, err %v", got, err)
	}
}

func TestListContacts_WithLeadID(t *testing.T) {
	leadID := int64(42)
	s := &Service{repo: &mockRepo{contacts: []domain.Contact{{ID: 7, LeadID: 42}}}}
	got, err := s.ListContacts(context.Background(), 1, &leadID, 0, 0)
	if err != nil || len(got.Contacts) != 1 {
		t.Errorf("got %v, err %v", got, err)
	}
}

func TestCreateContact_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{contact: domain.Contact{ID: 2}}, ownerOf: fixedBusinessOwner{ownerID: 11}}
	got, err := s.CreateContact(context.Background(), 1, domain.CreateContactRequest{Name: "Alice"})
	if err != nil || got.ID != 2 {
		t.Errorf("got %+v, err %v", got, err)
	}
}

func TestDeleteContact_PropagatesError(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{err: want}}
	if err := s.DeleteContact(context.Background(), 1, 10); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// --- Touchpoint tests ---

func TestListTouchpoints_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{touchpoints: []domain.Touchpoint{{ID: 1}}}}
	got, err := s.ListTouchpoints(context.Background(), 1, 5)
	if err != nil || len(got) != 1 {
		t.Errorf("got %v, err %v", got, err)
	}
}

func TestCreateTouchpoint_AppliesDefaults(t *testing.T) {
	m := &mockRepo{touchpoint: domain.Touchpoint{ID: 1, Type: "call", Outcome: "other"}}
	s := &Service{repo: m}
	got, err := s.CreateTouchpoint(context.Background(), 1, 2, domain.CreateTouchpointRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Type != "call" || got.Outcome != "other" {
		t.Errorf("unexpected defaults: type=%q outcome=%q", got.Type, got.Outcome)
	}
}

func TestDeleteTouchpoint_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteTouchpoint(context.Background(), 1, 10); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Contract tests ---

func TestListContracts_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{contracts: []domain.Contract{{ID: 1, Plan: "pro"}}}}
	got, err := s.ListContracts(context.Background(), 1, 0, 0)
	if err != nil || len(got.Contracts) != 1 {
		t.Errorf("got %v, err %v", got, err)
	}
}

func TestGetContract_PropagatesError(t *testing.T) {
	want := domain.ErrContractNotFound
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.GetContract(context.Background(), 99, 1)
	if !errors.Is(err, want) {
		t.Errorf("expected ErrContractNotFound, got %v", err)
	}
}

func TestCreateContract_AppliesDefaults(t *testing.T) {
	m := &mockRepo{contract: domain.Contract{ID: 1, Plan: "pro", Seats: 1}}
	s := &Service{repo: m}
	got, err := s.CreateContract(context.Background(), 1, domain.CreateContractRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Plan != "pro" || got.Seats != 1 {
		t.Errorf("expected pro/1 defaults, got plan=%q seats=%d", got.Plan, got.Seats)
	}
}

func TestUpdateContract_PropagatesError(t *testing.T) {
	want := errors.New("update failed")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.UpdateContract(context.Background(), 1, 10, domain.UpdateContractRequest{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func (m *mockRepo) ListSignedContracts(_ context.Context, _ int64, _, _ int32) ([]domain.Contract, error) {
	return m.contracts, m.err
}

func (m *mockRepo) CountSignedContracts(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.contracts)), m.err
}

func (m *mockRepo) ListUnsignedContracts(_ context.Context, _ int64, _, _ int32) ([]domain.Contract, error) {
	return m.contracts, m.err
}

func (m *mockRepo) CountUnsignedContracts(_ context.Context, _ int64) (int64, error) {
	return int64(len(m.contracts)), m.err
}
