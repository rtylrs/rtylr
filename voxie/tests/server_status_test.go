package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

type generatedStatusError struct{}

func (generatedStatusError) Error() string { return "business access denied" }

func (generatedStatusError) HTTPStatusCode() int { return http.StatusForbidden }

type statusAwareUserServer struct{}

func (statusAwareUserServer) GetMe(context.Context, *v1.GetMeRequest) (*v1.GetMeResponse, error) {
	return &v1.GetMeResponse{}, nil
}

func (statusAwareUserServer) UpdateProfile(context.Context, *v1.UpdateProfileRequest) (*v1.UpdateProfileResponse, error) {
	return &v1.UpdateProfileResponse{}, nil
}

func (statusAwareUserServer) ChangePassword(context.Context, *v1.ChangePasswordRequest) (*v1.ChangePasswordResponse, error) {
	return &v1.ChangePasswordResponse{}, nil
}

func (statusAwareUserServer) SwitchBusiness(context.Context, *v1.SwitchBusinessRequest) (*v1.SwitchBusinessResponse, error) {
	return nil, generatedStatusError{}
}

func TestGeneratedServerPreservesOneKitHTTPStatus(t *testing.T) {
	mux := http.NewServeMux()
	if err := v1.RegisterUserServiceServer(statusAwareUserServer{}, v1.WithMux(mux)); err != nil {
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
}
