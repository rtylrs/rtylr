package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/documents/internal/infrastructure"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/hr/documents/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hrdocumentv1 "github.com/voxire/voxie/gen/go/hr/document/v1"
)

func TestCreateDocument_ValidatesStaffAgainstBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const (
		businessID = int64(993015)
		userID     = int64(994015)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)
	staffResult, err := db.ExecContext(ctx, `
		INSERT INTO hr_staff
			(owner_id, business_id, first_name, last_name, email, role, start_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		businessID, businessID, "Scoped", "Staff", "documents-round3@example.test", "staff", "2026-01-01",
	)
	if err != nil {
		t.Fatalf("seed staff: %v", err)
	}
	staffID, err := staffResult.LastInsertId()
	if err != nil {
		t.Fatalf("read staff id: %v", err)
	}
	foreignStaffResult, err := db.ExecContext(ctx, `
		INSERT INTO hr_staff
			(owner_id, business_id, first_name, last_name, email, role, start_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		businessID+1, businessID+1, "Foreign", "Staff", "documents-round3-foreign@example.test", "staff", "2026-01-01",
	)
	if err != nil {
		t.Fatalf("seed foreign staff: %v", err)
	}
	foreignStaffID, err := foreignStaffResult.LastInsertId()
	if err != nil {
		t.Fatalf("read foreign staff id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM hr_staff WHERE id IN (?, ?)`, staffID, foreignStaffID)
	})

	api := NewDocumentAPI(application.NewService(ctx, moduleinfra.NewRepository(db)))
	authCtx := documentAPIAuthContext(businessID, userID, "hr:employees:view", "hr:employees:manage")
	for _, test := range []struct {
		name    string
		staffID int64
	}{
		{name: "foreign", staffID: foreignStaffID},
		{name: "missing", staffID: 999999999},
	} {
		t.Run(test.name, func(t *testing.T) {
			documentName := "Rejected " + test.name + " staff document"
			if _, err := api.CreateDocument(authCtx, &hrdocumentv1.CreateDocumentRequest{
				StaffId: test.staffID,
				Name:    documentName,
			}); documentErrorStatus(t, err) != http.StatusNotFound {
				t.Fatalf("CreateDocument error = %v, want HTTP 404", err)
			}

			var count int
			if err := db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM hr_documents WHERE business_id = ? AND name = ?`,
				businessID, documentName,
			).Scan(&count); err != nil {
				t.Fatalf("count rejected document: %v", err)
			}
			if count != 0 {
				t.Fatalf("rejected staff association persisted %d document(s)", count)
			}
		})
	}

	created, err := api.CreateDocument(authCtx, &hrdocumentv1.CreateDocumentRequest{
		StaffId: staffID,
		Name:    "Valid business staff document",
	})
	if err != nil {
		t.Fatalf("CreateDocument for staff in business: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			`DELETE FROM hr_documents WHERE id = ? AND business_id = ?`, created.Id, businessID)
	})

	var persistedStaffID int64
	if err := db.QueryRowContext(ctx,
		`SELECT staff_id FROM hr_documents WHERE id = ? AND business_id = ?`,
		created.Id, businessID,
	).Scan(&persistedStaffID); err != nil {
		t.Fatalf("load created staff association: %v", err)
	}
	if persistedStaffID != staffID {
		t.Fatalf("persisted staff_id = %d, want %d", persistedStaffID, staffID)
	}
}

func TestDeleteDocument_MissingReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const (
		businessID = int64(993016)
		userID     = int64(994016)
	)
	testutil.GrantBusinessAccess(t, db, businessID, userID)

	api := NewDocumentAPI(application.NewService(ctx, moduleinfra.NewRepository(db)))
	_, err := api.DeleteDocument(
		documentAPIAuthContext(businessID, userID, "hr:employees:view", "hr:employees:manage"),
		&hrdocumentv1.DeleteDocumentRequest{Id: 999999999},
	)
	if documentErrorStatus(t, err) != http.StatusNotFound {
		t.Fatalf("DeleteDocument error = %v, want HTTP 404", err)
	}
}

func TestDocumentAPIRequiresCanonicalEmployeePermissions(t *testing.T) {
	api := NewDocumentAPI(nil)
	viewCtx := documentAPIAuthContext(1, 1, "hr:employees:view")
	manageCtx := documentAPIAuthContext(1, 1, "hr:employees:manage")

	if _, err := api.ListDocuments(manageCtx, &hrdocumentv1.ListDocumentsRequest{}); documentErrorStatus(t, err) != http.StatusForbidden {
		t.Fatalf("ListDocuments without view permission error = %v, want HTTP 403", err)
	}
	if _, err := api.CreateDocument(viewCtx, &hrdocumentv1.CreateDocumentRequest{Name: "Denied"}); documentErrorStatus(t, err) != http.StatusForbidden {
		t.Fatalf("CreateDocument without manage permission error = %v, want HTTP 403", err)
	}
	if _, err := api.DeleteDocument(viewCtx, &hrdocumentv1.DeleteDocumentRequest{Id: 1}); documentErrorStatus(t, err) != http.StatusForbidden {
		t.Fatalf("DeleteDocument without manage permission error = %v, want HTTP 403", err)
	}
}

func documentAPIAuthContext(businessID, userID int64, permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: businessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: businessID,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	})
}

func documentErrorStatus(t *testing.T, err error) int {
	t.Helper()
	var httpErr *onekithttp.Error
	if errors.As(err, &httpErr) {
		return int(httpErr.GetStatusCode())
	}
	return 0
}
