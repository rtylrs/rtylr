package errors

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestServerError_Error(t *testing.T) {
	e := ServerError{Message: "boom"}
	if e.Error() != "boom" {
		t.Fatalf("got %q", e.Error())
	}
}

func TestServerError_MarshalJSON_OmitsTransientFields(t *testing.T) {
	loc := "/login"
	proxy := "http://internal/api"
	e := ServerError{
		Code:             400,
		HTTPCode:         400,
		Message:          "bad request",
		ProxyReRoute:     &proxy,
		RedirectLocation: &loc,
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(b)
	for _, must := range []string{"\"code\":400", "\"message\":\"bad request\""} {
		if !strings.Contains(s, must) {
			t.Errorf("expected %q in %s", must, s)
		}
	}
	for _, mustNot := range []string{"ProxyReRoute", "RedirectLocation", "/login", "internal/api"} {
		if strings.Contains(s, mustNot) {
			t.Errorf("unexpected %q in %s", mustNot, s)
		}
	}
}

func TestPermissionDeniedServerError(t *testing.T) {
	e := PermissionDeniedServerError()
	if e == nil || e.Code != http.StatusForbidden || e.HTTPCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %+v", e)
	}
	if e.Message != ErrorPermission {
		t.Errorf("expected message %q, got %q", ErrorPermission, e.Message)
	}
}

func TestRedirectServerError(t *testing.T) {
	e := RedirectServerError("/somewhere")
	if e == nil || e.RedirectLocation == nil || *e.RedirectLocation != "/somewhere" {
		t.Fatalf("expected redirect to /somewhere, got %+v", e)
	}
}

func TestErrorResponse_GetDialog(t *testing.T) {
	withDialog := ErrorResponse{Error: ServerError{ErrorDialog: &Dialog{Title: "T", Message: "M"}}}
	d, ok := withDialog.GetDialog()
	if !ok || d.Title != "T" || d.Message != "M" {
		t.Errorf("expected dialog with T/M, got %+v ok=%v", d, ok)
	}

	withoutDialog := ErrorResponse{Error: ServerError{}}
	if _, ok := withoutDialog.GetDialog(); ok {
		t.Error("expected ok=false when no dialog")
	}
}

func TestDialog_ToOutput(t *testing.T) {
	d := Dialog{Title: "x", Message: "y"}
	out := d.ToOutput()
	if out.Title != "x" || out.Message != "y" {
		t.Errorf("got %+v", out)
	}
}
