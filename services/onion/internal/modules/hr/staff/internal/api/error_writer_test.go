package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	hrstaffv1 "github.com/voxire/voxie/gen/go/hr/staff/v1"
)

// The generated server only forwards an error's PublicCode through the
// ErrorWriter; its default body is message-only. This drives the real
// generated path so a regression there drops the code visibly.
func TestWriteErrorCarriesDirectoryReasonCode(t *testing.T) {
	cases := map[string]struct {
		err  error
		code string
	}{
		"disabled":           {errDirectoryDisabled, directoryDisabledCode},
		"missing permission": {errDirectoryPermissionMissing, directoryPermissionMissingCode},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/hr/staff/v1/staff", nil)
			hrstaffv1.ServerOptions{ErrorWriter: WriteError}.WriteHandlerError(rec, req, tc.err)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", rec.Code)
			}
			var body struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Code != tc.code || body.Message == "" {
				t.Fatalf("body = %+v, want code %q and a message", body, tc.code)
			}
		})
	}
}
