package handlers

import (
	"context"
	"testing"
	"time"
)

type fakeLoginRecorder struct {
	recorded chan error
}

func (f *fakeLoginRecorder) RecordLogin(ctx context.Context, userID int64, email, method, ip string, success bool) error {
	if userID != 42 || email != "alice@example.com" || method != "email" || ip != "127.0.0.1" || !success {
		f.recorded <- context.Canceled
		return nil
	}
	f.recorded <- ctx.Err()
	return nil
}

func TestRunAsyncUserHookIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	done := make(chan error, 1)
	runAsyncUserHook(parent, 42, func(ctx context.Context, userID int64) {
		if userID != 42 {
			done <- context.Canceled
			return
		}
		select {
		case <-ctx.Done():
			done <- ctx.Err()
		default:
			done <- nil
		}
	})

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected hook context to survive parent cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async hook")
	}
}

func TestRunAsyncLoginTasksIgnoresParentCancellation(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	recorder := &fakeLoginRecorder{recorded: make(chan error, 1)}
	hookDone := make(chan error, 1)
	runAsyncLoginTasks(parent, recorder, 42, "alice@example.com", "127.0.0.1", func(ctx context.Context, userID int64) {
		if userID != 42 {
			hookDone <- context.Canceled
			return
		}
		hookDone <- ctx.Err()
	})

	select {
	case err := <-recorder.recorded:
		if err != nil {
			t.Fatalf("expected login record context to survive parent cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for login recorder")
	}

	select {
	case err := <-hookDone:
		if err != nil {
			t.Fatalf("expected login hook context to survive parent cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for login hook")
	}
}

func TestLoginHandler_Validate_RequiresEmail(t *testing.T) {
	h := &LoginHandler{}
	req := &LoginRequest{Email: "", Password: "secret123"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when email is empty")
	}
}

func TestLoginHandler_Validate_RequiresPassword(t *testing.T) {
	h := &LoginHandler{}
	req := &LoginRequest{Email: "alice@example.com", Password: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when password is empty")
	}
}

func TestLoginHandler_Validate_HappyPath(t *testing.T) {
	h := &LoginHandler{}
	req := &LoginRequest{Email: "alice@example.com", Password: "hunter2"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoginHandler_Validate_BothEmpty(t *testing.T) {
	h := &LoginHandler{}
	req := &LoginRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both fields are empty")
	}
}

func TestChangePasswordHandler_Validate_RequiresCurrentPassword(t *testing.T) {
	h := &ChangePasswordHandler{}
	req := &ChangePasswordRequest{CurrentPassword: "", NewPassword: "newpass123"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when current password is empty")
	}
}

func TestChangePasswordHandler_Validate_RequiresNewPassword(t *testing.T) {
	h := &ChangePasswordHandler{}
	req := &ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when new password is empty")
	}
}

func TestChangePasswordHandler_Validate_HappyPath(t *testing.T) {
	h := &ChangePasswordHandler{}
	req := &ChangePasswordRequest{CurrentPassword: "oldpass", NewPassword: "newpass123"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChangePasswordHandler_Validate_BothEmpty(t *testing.T) {
	h := &ChangePasswordHandler{}
	req := &ChangePasswordRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both fields are empty")
	}
}

func TestDeleteAccountHandler_Validate_RequiresPassword(t *testing.T) {
	h := &DeleteAccountHandler{}
	req := &DeleteAccountRequest{Password: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when password is empty")
	}
}

func TestDeleteAccountHandler_Validate_HappyPath(t *testing.T) {
	h := &DeleteAccountHandler{}
	req := &DeleteAccountRequest{Password: "mypassword"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateProfileHandler_Validate_RequiresName(t *testing.T) {
	h := &UpdateProfileHandler{}
	req := &UpdateProfileRequest{Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when name is empty")
	}
}

func TestUpdateProfileHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateProfileHandler{}
	req := &UpdateProfileRequest{Name: "Alice Smith"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordResetRequestHandler_Validate_RequiresEmail(t *testing.T) {
	h := &PasswordResetRequestHandler{}
	req := &PasswordResetRequestBody{Email: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when email is empty")
	}
}

func TestPasswordResetRequestHandler_Validate_HappyPath(t *testing.T) {
	h := &PasswordResetRequestHandler{}
	req := &PasswordResetRequestBody{Email: "bob@example.com"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordResetConfirmHandler_Validate_RequiresToken(t *testing.T) {
	h := &PasswordResetConfirmHandler{}
	req := &PasswordResetConfirmBody{Token: "", Password: "newpass"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when token is empty")
	}
}

func TestPasswordResetConfirmHandler_Validate_RequiresPassword(t *testing.T) {
	h := &PasswordResetConfirmHandler{}
	req := &PasswordResetConfirmBody{Token: "abc123", Password: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when password is empty")
	}
}

func TestPasswordResetConfirmHandler_Validate_HappyPath(t *testing.T) {
	h := &PasswordResetConfirmHandler{}
	req := &PasswordResetConfirmBody{Token: "abc123xyz", Password: "newpassword"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
