package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

// fakeProjects answers ProjectExists from a business -> project id set.
type fakeProjects struct {
	byBusiness map[int64][]int64
	err        error
	calls      int
}

func (f *fakeProjects) ProjectExists(_ context.Context, businessID, projectID int64) (bool, error) {
	f.calls++
	if f.err != nil {
		return false, f.err
	}
	for _, id := range f.byBusiness[businessID] {
		if id == projectID {
			return true, nil
		}
	}
	return false, nil
}

func projectInvoiceInput(projectID int64) domain.InvoiceInput {
	return domain.InvoiceInput{
		ClientName: "Acme",
		ProjectID:  projectID,
		IssueDate:  "2024-06-15",
		DueDate:    "2024-06-30",
		Items:      []domain.InvoiceItemInput{{Description: "Design", Quantity: 1, UnitPrice: 100}},
	}
}

func projectUpdateInput(projectID *int64) domain.InvoiceUpdateInput {
	return domain.InvoiceUpdateInput{
		ClientName: "Acme",
		ProjectID:  projectID,
		IssueDate:  "2024-06-15",
		DueDate:    "2024-06-30",
		Items:      []domain.InvoiceItemInput{{Description: "Design", Quantity: 1, UnitPrice: 100}},
	}
}

func TestCreateInvoice_StoresVerifiedProject(t *testing.T) {
	repo := &mockRepo{generatedID: 9}
	projects := &fakeProjects{byBusiness: map[int64][]int64{1: {55}}}
	s := &Service{repo: repo, projectReferences: projects}

	inv, err := s.CreateInvoice(context.Background(), 1, projectInvoiceInput(55))
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if repo.created == nil || repo.created.ProjectID == nil || *repo.created.ProjectID != 55 {
		t.Fatalf("persisted project = %v, want 55", repo.created)
	}
	if inv.ProjectID == nil || *inv.ProjectID != 55 {
		t.Fatalf("returned project = %v, want 55", inv.ProjectID)
	}
}

func TestCreateInvoice_RejectsProjectFromAnotherBusiness(t *testing.T) {
	repo := &mockRepo{generatedID: 9}
	projects := &fakeProjects{byBusiness: map[int64][]int64{2: {55}}}
	s := &Service{repo: repo, projectReferences: projects}

	_, err := s.CreateInvoice(context.Background(), 1, projectInvoiceInput(55))
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("error = %v, want ErrProjectNotFound", err)
	}
	if repo.createWithItemsCalls != 0 {
		t.Fatalf("an invoice was written for a foreign project")
	}
}

func TestCreateInvoice_WithoutProjectSkipsVerification(t *testing.T) {
	repo := &mockRepo{generatedID: 9}
	s := &Service{repo: repo}

	if _, err := s.CreateInvoice(context.Background(), 1, projectInvoiceInput(0)); err != nil {
		t.Fatalf("CreateInvoice without a project: %v", err)
	}
	if repo.created == nil || repo.created.ProjectID != nil {
		t.Fatalf("untagged invoice persisted a project: %+v", repo.created)
	}
}

func TestCreateInvoice_ProjectFailsClosedWhenUnwired(t *testing.T) {
	repo := &mockRepo{generatedID: 9}
	s := &Service{repo: repo}

	_, err := s.CreateInvoice(context.Background(), 1, projectInvoiceInput(55))
	if !errors.Is(err, ErrProjectVerificationUnavailable) {
		t.Fatalf("error = %v, want ErrProjectVerificationUnavailable", err)
	}
	if repo.createWithItemsCalls != 0 {
		t.Fatalf("an invoice was written with an unverified project")
	}
}

func TestCreateInvoice_ProjectLookupErrorIsReturned(t *testing.T) {
	want := errors.New("lookup failed")
	repo := &mockRepo{generatedID: 9}
	s := &Service{repo: repo, projectReferences: &fakeProjects{err: want}}

	if _, err := s.CreateInvoice(context.Background(), 1, projectInvoiceInput(55)); !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestUpdateInvoice_ProjectTagStates(t *testing.T) {
	existingProject := int64(55)
	zero := int64(0)
	other := int64(77)
	foreign := int64(88)

	tests := []struct {
		name    string
		input   *int64
		want    *int64
		wantErr error
	}{
		{name: "absent keeps the tag", input: nil, want: &existingProject},
		{name: "zero clears the tag", input: &zero, want: nil},
		{name: "new id re-tags", input: &other, want: &other},
		{name: "foreign id is rejected", input: &foreign, wantErr: ErrProjectNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tagged := existingProject
			repo := &mockRepo{invoice: &domain.Invoice{ID: 7, Status: domain.StatusDraft, ProjectID: &tagged}}
			projects := &fakeProjects{byBusiness: map[int64][]int64{1: {55, 77}, 2: {88}}}
			s := &Service{repo: repo, projectReferences: projects}

			inv, err := s.UpdateInvoice(context.Background(), 1, 7, projectUpdateInput(tc.input))
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
				if repo.updateWithItemsCalls != 0 {
					t.Fatalf("invoice was written despite a rejected project")
				}
				return
			}
			if err != nil {
				t.Fatalf("UpdateInvoice: %v", err)
			}
			switch {
			case tc.want == nil && inv.ProjectID != nil:
				t.Fatalf("project = %d, want none", *inv.ProjectID)
			case tc.want != nil && (inv.ProjectID == nil || *inv.ProjectID != *tc.want):
				t.Fatalf("project = %v, want %d", inv.ProjectID, *tc.want)
			}
			if tc.input == nil && projects.calls != 0 {
				t.Fatalf("an absent project was verified %d times", projects.calls)
			}
		})
	}
}
