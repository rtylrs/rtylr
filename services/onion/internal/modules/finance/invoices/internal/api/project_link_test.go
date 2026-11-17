package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

// stubProjectReferences stands in for flow's reference reader: only the listed
// project ids exist in the caller's business.
type stubProjectReferences map[int64]bool

func (s stubProjectReferences) ProjectExists(_ context.Context, _, projectID int64) (bool, error) {
	return s[projectID], nil
}

func newProjectLinkAPI(t *testing.T, businessID, userID int64, reader application.ProjectReferenceReader) (*InvoicesAPI, context.Context) {
	t.Helper()

	db := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, db, businessID, userID)
	t.Cleanup(func() {
		ctx := context.Background()
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id IN (SELECT id FROM finance_invoices WHERE business_id = ?)", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE business_id = ?", businessID)
	})

	user := financeUser("finance:invoices:create", "finance:invoices:view")
	user.BusinessID = businessID
	user.UserID = userID
	ctx := authPkg.WithAuthUser(context.Background(), user)

	svc := newInvoicesServiceForTest(db)
	svc.SetProjectReferenceReader(reader)
	return NewInvoicesAPI(svc), ctx
}

func projectInvoiceRequest(projectID *int64) *financeinvoicesv1.CreateInvoiceRequest {
	return &financeinvoicesv1.CreateInvoiceRequest{
		ClientName:   "Project Client",
		ClientEmail:  "billing@project.test",
		ProjectId:    projectID,
		IssueDate:    "2026-06-15",
		DueDate:      "2026-06-30",
		Currency:     "USD",
		DiscountType: "fixed",
		Items: []*financeinvoicesv1.InvoiceItem{
			{Description: "Consulting", Quantity: 1, UnitPrice: 100},
		},
	}
}

func updateBodyFrom(invoice *financeinvoicesv1.InvoiceDetail, projectID *int64) *financeinvoicesv1.UpdateInvoiceBody {
	return &financeinvoicesv1.UpdateInvoiceBody{
		ClientName:   invoice.GetClientName(),
		ClientEmail:  invoice.GetClientEmail(),
		ProjectId:    projectID,
		IssueDate:    invoice.GetIssueDate(),
		DueDate:      invoice.GetDueDate(),
		Currency:     invoice.GetCurrency(),
		DiscountType: "fixed",
		Items: []*financeinvoicesv1.InvoiceItem{
			{Description: "Consulting", Quantity: 1, UnitPrice: 100},
		},
	}
}

func TestInvoiceProjectLinkThroughAPI(t *testing.T) {
	const businessID int64 = 993201
	const userID int64 = 993201
	projectID := int64(6201)
	apiSvc, ctx := newProjectLinkAPI(t, businessID, userID, stubProjectReferences{projectID: true})

	created, err := apiSvc.CreateInvoice(ctx, projectInvoiceRequest(&projectID))
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	invoice := created.GetInvoice()
	if invoice.GetProjectId() != projectID {
		t.Fatalf("create response project_id = %d, want %d", invoice.GetProjectId(), projectID)
	}
	if _, err := apiSvc.CreateInvoice(ctx, projectInvoiceRequest(nil)); err != nil {
		t.Fatalf("CreateInvoice(untagged): %v", err)
	}

	listed, err := apiSvc.ListInvoices(ctx, &financeinvoicesv1.ListInvoicesRequest{ProjectId: &projectID})
	if err != nil {
		t.Fatalf("ListInvoices(project_id): %v", err)
	}
	if len(listed.GetInvoices()) != 1 || listed.GetInvoices()[0].GetId() != invoice.GetId() {
		t.Fatalf("ListInvoices(project_id=%d) = %+v, want only invoice %d", projectID, listed.GetInvoices(), invoice.GetId())
	}
	if listed.GetInvoices()[0].GetProjectId() != projectID {
		t.Fatalf("summary project_id = %d, want %d", listed.GetInvoices()[0].GetProjectId(), projectID)
	}

	// An update that omits project_id keeps the tag.
	kept, err := apiSvc.UpdateInvoice(ctx, &financeinvoicesv1.UpdateInvoiceRequest{Id: invoice.GetId(), Body: updateBodyFrom(invoice, nil)})
	if err != nil {
		t.Fatalf("UpdateInvoice(absent project): %v", err)
	}
	if kept.GetInvoice().GetProjectId() != projectID {
		t.Fatalf("project after absent update = %d, want %d", kept.GetInvoice().GetProjectId(), projectID)
	}

	// An explicit 0 clears it, and the detail read agrees.
	zero := int64(0)
	if _, err := apiSvc.UpdateInvoice(ctx, &financeinvoicesv1.UpdateInvoiceRequest{Id: invoice.GetId(), Body: updateBodyFrom(invoice, &zero)}); err != nil {
		t.Fatalf("UpdateInvoice(clear project): %v", err)
	}
	detail, err := apiSvc.GetInvoice(ctx, &financeinvoicesv1.GetInvoiceRequest{Id: invoice.GetId()})
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	if detail.GetInvoice().ProjectId != nil {
		t.Fatalf("detail project_id = %d after clear, want absent", *detail.GetInvoice().ProjectId)
	}
}

func TestCreateInvoiceUnknownProjectIsNotFound(t *testing.T) {
	const businessID int64 = 993202
	const userID int64 = 993202
	apiSvc, ctx := newProjectLinkAPI(t, businessID, userID, stubProjectReferences{})

	foreign := int64(6202)
	_, err := apiSvc.CreateInvoice(ctx, projectInvoiceRequest(&foreign))
	if got := financePermissionStatus(err); got != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (err = %v)", got, http.StatusNotFound, err)
	}
}
