package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type generatedStatusError struct{}

func (generatedStatusError) Error() string { return "business access denied" }

func (generatedStatusError) HTTPStatusCode() int { return http.StatusForbidden }

type statusAwareUserServer struct{}

func (statusAwareUserServer) GetMe(context.Context, *GetMeRequest) (*GetMeResponse, error) {
	return &GetMeResponse{}, nil
}

func (statusAwareUserServer) UpdateProfile(context.Context, *UpdateProfileRequest) (*UpdateProfileResponse, error) {
	return &UpdateProfileResponse{}, nil
}

func (statusAwareUserServer) ChangePassword(context.Context, *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	return &ChangePasswordResponse{}, nil
}

func (statusAwareUserServer) SwitchBusiness(context.Context, *SwitchBusinessRequest) (*SwitchBusinessResponse, error) {
	return nil, generatedStatusError{}
}

func TestGeneratedServerPreservesOneKitHTTPStatus(t *testing.T) {
	mux := http.NewServeMux()
	if err := RegisterUserServiceServer(statusAwareUserServer{}, WithMux(mux)); err != nil {
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
