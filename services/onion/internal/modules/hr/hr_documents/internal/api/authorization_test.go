package api

import (
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
	"context"
	"net/http"
	"testing"


	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

func documentAuthContext(permissions ...string) context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     1001,
		BusinessID: 1001,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
		},
	})
}

func documentIDContext(ctx context.Context) context.Context {
	routeCtx := map[string]string{}
	routeCtx["id"] = ("1")
	return serverHttp.WithPathParams(ctx, routeCtx)
}

func TestDocumentHandlersRequireCanonicalPermissions(t *testing.T) {
	ctx := documentAuthContext("hr:time_off:request")
	tests := []struct {
		name string
		call func(context.Context) error
	}{
		{
			name: "all employee list",
			call: func(ctx context.Context) error {
				_, err := (&ListDocumentsHandler{}).Serve(ctx, &ListDocumentsRequest{})
				return err
			},
		},
		{
			name: "employee list",
			call: func(ctx context.Context) error {
				_, err := (&ListDocumentsHandler{}).Serve(ctx, &ListDocumentsRequest{EmployeeID: 4101})
				return err
			},
		},
		{
			name: "expiring list",
			call: func(ctx context.Context) error {
				_, err := (&ListExpiringDocumentsHandler{}).Serve(ctx, &ListExpiringDocumentsRequest{})
				return err
			},
		},
		{
			name: "get",
			call: func(ctx context.Context) error {
				_, err := (&GetDocumentHandler{}).Serve(documentIDContext(ctx), &GetDocumentRequest{})
				return err
			},
		},
		{
			name: "upload",
			call: func(ctx context.Context) error {
				_, err := (&UploadDocumentHandler{}).Serve(ctx, &UploadDocumentRequest{
					EmployeeID:  4101,
					Title:       "Contract",
					StorageMode: "stored",
					FileURL:     "https://example.com/contract.pdf",
				})
				return err
			},
		},
		{
			name: "delete",
			call: func(ctx context.Context) error {
				_, err := (&DeleteDocumentHandler{}).Serve(documentIDContext(ctx), &DeleteDocumentRequest{})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if status := oneKitStatus(test.call(ctx)); status != http.StatusForbidden {
				t.Fatalf("expected 403 without employee permissions, got %d", status)
			}
		})
	}
}

func TestDocumentMutationRequiresManageInsteadOfView(t *testing.T) {
	ctx := documentAuthContext("hr:employees:view")
	_, err := (&UploadDocumentHandler{}).Serve(ctx, &UploadDocumentRequest{
		EmployeeID:  4101,
		Title:       "Contract",
		StorageMode: "stored",
		FileURL:     "https://example.com/contract.pdf",
	})
	if status := oneKitStatus(err); status != http.StatusForbidden {
		t.Fatalf("expected view-only user to receive 403 for upload, got %d", status)
	}
}
