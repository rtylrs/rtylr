package onekiterrors_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

type statusAwareUserServer struct{}

func (statusAwareUserServer) GetMe(context.Context, *hubuserv1.GetMeRequest) (*hubuserv1.GetMeResponse, error) {
	return &hubuserv1.GetMeResponse{}, nil
}

func (statusAwareUserServer) UpdateProfile(context.Context, *hubuserv1.UpdateProfileRequest) (*hubuserv1.UpdateProfileResponse, error) {
	return &hubuserv1.UpdateProfileResponse{}, nil
}

func (statusAwareUserServer) ChangePassword(context.Context, *hubuserv1.ChangePasswordRequest) (*hubuserv1.ChangePasswordResponse, error) {
	return &hubuserv1.ChangePasswordResponse{}, nil
}

func (statusAwareUserServer) SwitchBusiness(context.Context, *hubuserv1.SwitchBusinessRequest) (*hubuserv1.SwitchBusinessResponse, error) {
	return nil, onekiterrors.New(http.StatusForbidden, "business access denied")
}

func TestGeneratedServerPreservesOneKitHTTPStatus(t *testing.T) {
	mux := http.NewServeMux()
	if err := hubuserv1.RegisterUserServiceServer(statusAwareUserServer{}, hubuserv1.WithMux(mux)); err != nil {
		t.Fatalf("register user service: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/hub/user/v1/users/switch-business",
		strings.NewReader(`{"business_id":"1"}`),
	)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected generated server to preserve 403, got %d with body %q", res.Code, res.Body.String())
	}
	if body := res.Body.String(); !strings.Contains(body, "business access denied") {
		t.Fatalf("expected generated server to preserve the handler message, got body %q", body)
	}
}
