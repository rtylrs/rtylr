package api

import (
	"context"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

// invoiceMoney is everything a project re-tag must leave alone.
type invoiceMoney struct {
	status                                 string
	subtotal, tax, total, amountPaid, owed float64
	items                                  int
}

func moneyOf(inv *financeinvoicesv1.InvoiceDetail) invoiceMoney {
	return invoiceMoney{
		status:     inv.GetStatus(),
		subtotal:   inv.GetSubtotal(),
		tax:        inv.GetTaxAmount(),
		total:      inv.GetTotal(),
		amountPaid: inv.GetAmountPaid(),
		owed:       inv.GetAmountDue(),
		items:      len(inv.GetItems()),
	}
}

func setProject(id, projectID int64) *financeinvoicesv1.SetInvoiceProjectRequest {
	return &financeinvoicesv1.SetInvoiceProjectRequest{Id: id, Body: &financeinvoicesv1.SetInvoiceProjectBody{ProjectId: &projectID}}
}

func TestSetInvoiceProjectOnIssuedInvoices(t *testing.T) {
	const businessID int64 = 993211
	const userID int64 = 993211
	project := int64(6211)
	other := int64(6212)
	foreign := int64(6213)
	apiSvc, ctx := newProjectLinkAPI(t, businessID, userID, stubProjectReferences{project: true, other: true})
	db := testutil.OpenDB(t)

	created, err := apiSvc.CreateInvoice(ctx, projectInvoiceRequest(nil))
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	id := created.GetInvoice().GetId()
	if _, err := apiSvc.UpdateInvoiceStatus(ctx, &financeinvoicesv1.UpdateInvoiceStatusRequest{Id: id, Body: &financeinvoicesv1.UpdateInvoiceStatusBody{Status: "sent"}}); err != nil {
		t.Fatalf("mark sent: %v", err)
	}
	before, err := apiSvc.GetInvoice(ctx, &financeinvoicesv1.GetInvoiceRequest{Id: id})
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	want := moneyOf(before.GetInvoice())
	if want.status != "sent" {
		t.Fatalf("status before = %q, want sent", want.status)
	}

	// Warm the per-project list so a stale cache would show up below.
	if listed, err := apiSvc.ListInvoices(ctx, &financeinvoicesv1.ListInvoicesRequest{ProjectId: &project}); err != nil || len(listed.GetInvoices()) != 0 {
		t.Fatalf("ListInvoices(project) before = %v, %v; want empty", listed.GetInvoices(), err)
	}

	// A sent invoice takes a project; nothing else moves.
	tagged, err := apiSvc.SetInvoiceProject(ctx, setProject(id, project))
	if err != nil {
		t.Fatalf("SetInvoiceProject(sent): %v", err)
	}
	if tagged.GetInvoice().GetProjectId() != project {
		t.Fatalf("response project_id = %d, want %d", tagged.GetInvoice().GetProjectId(), project)
	}
	if got := moneyOf(tagged.GetInvoice()); got != want {
		t.Fatalf("response totals/status = %+v, want %+v", got, want)
	}
	stored, err := apiSvc.GetInvoice(ctx, &financeinvoicesv1.GetInvoiceRequest{Id: id})
	if err != nil {
		t.Fatalf("GetInvoice after tag: %v", err)
	}
	if stored.GetInvoice().GetProjectId() != project || moneyOf(stored.GetInvoice()) != want {
		t.Fatalf("stored invoice = project %d %+v, want project %d %+v", stored.GetInvoice().GetProjectId(), moneyOf(stored.GetInvoice()), project, want)
	}
	listed, err := apiSvc.ListInvoices(ctx, &financeinvoicesv1.ListInvoicesRequest{ProjectId: &project})
	if err != nil || len(listed.GetInvoices()) != 1 || listed.GetInvoices()[0].GetId() != id {
		t.Fatalf("ListInvoices(project) after = %v, %v; want only invoice %d", listed.GetInvoices(), err, id)
	}

	// A project from another business is a 404 and keeps the tag.
	if _, err := apiSvc.SetInvoiceProject(ctx, setProject(id, foreign)); financePermissionStatus(err) != http.StatusNotFound {
		t.Fatalf("foreign project status = %d, want 404 (err = %v)", financePermissionStatus(err), err)
	}

	// 0 clears.
	cleared, err := apiSvc.SetInvoiceProject(ctx, setProject(id, 0))
	if err != nil {
		t.Fatalf("SetInvoiceProject(0): %v", err)
	}
	if cleared.GetInvoice().ProjectId != nil {
		t.Fatalf("project after clear = %d, want absent", *cleared.GetInvoice().ProjectId)
	}

	// Paid is terminal for money but still counts toward a project.
	if _, err := db.ExecContext(context.Background(), "UPDATE finance_invoices SET status = 'paid', amount_paid = total, amount_due = 0 WHERE id = ?", id); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	paid, err := apiSvc.SetInvoiceProject(ctx, setProject(id, other))
	if err != nil {
		t.Fatalf("SetInvoiceProject(paid): %v", err)
	}
	if paid.GetInvoice().GetProjectId() != other || paid.GetInvoice().GetStatus() != "paid" || paid.GetInvoice().GetAmountDue() != 0 {
		t.Fatalf("paid invoice after tag = %+v", paid.GetInvoice())
	}
}

func TestSetInvoiceProjectRefusesVoidInvoices(t *testing.T) {
	const businessID int64 = 993212
	const userID int64 = 993212
	project := int64(6221)
	apiSvc, ctx := newProjectLinkAPI(t, businessID, userID, stubProjectReferences{project: true})

	created, err := apiSvc.CreateInvoice(ctx, projectInvoiceRequest(&project))
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	id := created.GetInvoice().GetId()
	if _, err := apiSvc.UpdateInvoiceStatus(ctx, &financeinvoicesv1.UpdateInvoiceStatusRequest{Id: id, Body: &financeinvoicesv1.UpdateInvoiceStatusBody{Status: "void"}}); err != nil {
		t.Fatalf("void: %v", err)
	}

	_, err = apiSvc.SetInvoiceProject(ctx, setProject(id, 0))
	if got := financePermissionStatus(err); got != http.StatusConflict {
		t.Fatalf("status = %d, want %d (err = %v)", got, http.StatusConflict, err)
	}
	stored, err := apiSvc.GetInvoice(ctx, &financeinvoicesv1.GetInvoiceRequest{Id: id})
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	if stored.GetInvoice().GetProjectId() != project {
		t.Fatalf("void invoice project = %d, want %d unchanged", stored.GetInvoice().GetProjectId(), project)
	}
}

func TestSetInvoiceProjectNeedsAProjectID(t *testing.T) {
	const businessID int64 = 993213
	const userID int64 = 993213
	apiSvc, ctx := newProjectLinkAPI(t, businessID, userID, stubProjectReferences{})

	for name, req := range map[string]*financeinvoicesv1.SetInvoiceProjectRequest{
		"no body":       {Id: 1},
		"no project_id": {Id: 1, Body: &financeinvoicesv1.SetInvoiceProjectBody{}},
	} {
		if _, err := apiSvc.SetInvoiceProject(ctx, req); financePermissionStatus(err) != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400 (err = %v)", name, financePermissionStatus(err), err)
		}
	}
}

func TestSetInvoiceProjectIsScopedToTheBusiness(t *testing.T) {
	project := int64(6241)
	ownerAPI, ownerCtx := newProjectLinkAPI(t, 993214, 993214, stubProjectReferences{project: true})
	otherAPI, otherCtx := newProjectLinkAPI(t, 993215, 993215, stubProjectReferences{project: true})

	created, err := ownerAPI.CreateInvoice(ownerCtx, projectInvoiceRequest(&project))
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	id := created.GetInvoice().GetId()

	// Another business cannot see, let alone clear, the invoice.
	if _, err := otherAPI.SetInvoiceProject(otherCtx, setProject(id, 0)); financePermissionStatus(err) != http.StatusNotFound {
		t.Fatalf("cross-business status = %d, want 404 (err = %v)", financePermissionStatus(err), err)
	}
	if _, err := otherAPI.SetInvoiceProject(otherCtx, setProject(987654321, 0)); financePermissionStatus(err) != http.StatusNotFound {
		t.Fatalf("unknown invoice status = %d, want 404 (err = %v)", financePermissionStatus(err), err)
	}
	stored, err := ownerAPI.GetInvoice(ownerCtx, &financeinvoicesv1.GetInvoiceRequest{Id: id})
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	if stored.GetInvoice().GetProjectId() != project {
		t.Fatalf("project = %d after a cross-business clear, want %d", stored.GetInvoice().GetProjectId(), project)
	}
}

func TestSetInvoiceProjectRequiresInvoiceWritePermission(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), financeUser("finance:invoices:view", "finance:invoices:send"))

	_, err := (&InvoicesAPI{}).SetInvoiceProject(ctx, setProject(1, 0))
	if got := financePermissionStatus(err); got != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", got, http.StatusForbidden)
	}
}
