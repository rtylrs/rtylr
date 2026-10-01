package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/oauth/internal/domain"
	"github.com/voxire/onion/internal/modules/oauth/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetClient_UnknownClientReturnsErrClientNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetClient(ctx, "nonexistent-client-id")
	if !errors.Is(err, domain.ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}
}

func TestGetAuthCode_UnknownCodeReturnsErrCodeNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetAuthCode(ctx, "no-such-code")
	if !errors.Is(err, domain.ErrCodeNotFound) {
		t.Fatalf("expected ErrCodeNotFound, got %v", err)
	}
}

func TestGetTokenByRefresh_UnknownTokenReturnsErrTokenNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetTokenByRefresh(ctx, "no-such-refresh-token")
	if !errors.Is(err, domain.ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestSaveAndGetAuthCode_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	code := &domain.AuthCode{
		Code:                "test-auth-code-roundtrip",
		ClientID:            "test-client",
		UserID:              testutil.OwnerID,
		RedirectURI:         "https://example.com/callback",
		Scope:               "read",
		CodeChallenge:       "",
		CodeChallengeMethod: "plain",
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}

	if err := repo.SaveAuthCode(ctx, code); err != nil {
		t.Fatalf("SaveAuthCode error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM oauth_auth_codes WHERE code = ?", code.Code)
	})

	got, err := repo.GetAuthCode(ctx, code.Code)
	if err != nil {
		t.Fatalf("GetAuthCode error: %v", err)
	}
	if got.Code != code.Code {
		t.Errorf("code mismatch: got %q, want %q", got.Code, code.Code)
	}
	if got.UserID != code.UserID {
		t.Errorf("user_id mismatch: got %d, want %d", got.UserID, code.UserID)
	}
}

func TestMarkAuthCodeUsed_MarksRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	code := &domain.AuthCode{
		Code:                "test-auth-code-mark-used",
		ClientID:            "test-client",
		UserID:              testutil.OwnerID,
		RedirectURI:         "https://example.com/callback",
		Scope:               "read",
		CodeChallenge:       "",
		CodeChallengeMethod: "plain",
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	}
	if err := repo.SaveAuthCode(ctx, code); err != nil {
		t.Fatalf("SaveAuthCode error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM oauth_auth_codes WHERE code = ?", code.Code)
	})

	if err := repo.MarkAuthCodeUsed(ctx, code.Code); err != nil {
		t.Fatalf("MarkAuthCodeUsed error: %v", err)
	}

	got, err := repo.GetAuthCode(ctx, code.Code)
	if err != nil {
		t.Fatalf("GetAuthCode after marking used: %v", err)
	}
	if !got.Used {
		t.Error("expected Used=true after MarkAuthCodeUsed")
	}
}

func TestSaveAndGetToken_RoundTrip(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	token := &domain.Token{
		AccessToken:  "test-access-token-roundtrip",
		RefreshToken: "test-refresh-token-roundtrip",
		ClientID:     "test-client",
		UserID:       testutil.OwnerID,
		Scope:        "read",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	if err := repo.SaveToken(ctx, token); err != nil {
		t.Fatalf("SaveToken error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM oauth_tokens WHERE refresh_token = ?", token.RefreshToken)
	})

	got, err := repo.GetTokenByRefresh(ctx, token.RefreshToken)
	if err != nil {
		t.Fatalf("GetTokenByRefresh error: %v", err)
	}
	if got.RefreshToken != token.RefreshToken {
		t.Errorf("refresh_token mismatch: got %q, want %q", got.RefreshToken, token.RefreshToken)
	}
	if got.UserID != token.UserID {
		t.Errorf("user_id mismatch: got %d, want %d", got.UserID, token.UserID)
	}
}

func TestRevokeToken_SetsRevokedFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	token := &domain.Token{
		AccessToken:  "test-access-token-revoke",
		RefreshToken: "test-refresh-token-revoke",
		ClientID:     "test-client",
		UserID:       testutil.OwnerID,
		Scope:        "read",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := repo.SaveToken(ctx, token); err != nil {
		t.Fatalf("SaveToken error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM oauth_tokens WHERE refresh_token = ?", token.RefreshToken)
	})

	if err := repo.RevokeToken(ctx, token.RefreshToken); err != nil {
		t.Fatalf("RevokeToken error: %v", err)
	}

	got, err := repo.GetTokenByRefresh(ctx, token.RefreshToken)
	if err != nil {
		t.Fatalf("GetTokenByRefresh after revoke: %v", err)
	}
	if !got.Revoked {
		t.Error("expected Revoked=true after RevokeToken")
	}
}

func TestGetUserIDByEmail_UnknownEmailErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetUserIDByEmail(ctx, "does-not-exist@example.invalid")
	if err == nil {
		t.Fatal("expected error for unknown email, got nil")
	}
}

func TestGetUserPasswordHash_UnknownUserErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetUserPasswordHash(ctx, 0)
	if err == nil {
		t.Fatal("expected error for user_id=0, got nil")
	}
}

func TestGetUserBusinessID_UnknownUserReturnsErrUserNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	businessID, err := repo.GetUserBusinessID(ctx, 0)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if businessID != 0 {
		t.Errorf("expected business_id=0 for unknown user, got %d", businessID)
	}
}

func TestGetUserBusinessID_ReturnsSelectedBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	res, err := db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?)",
		"oauth-owner@example.com", "", "OAuth Owner",
	)
	if err != nil {
		t.Fatalf("insert owner: %v", err)
	}
	ownerID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("owner id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", ownerID)
	})

	res, err = db.ExecContext(ctx,
		"INSERT INTO businesses (name, owner_id) VALUES (?, ?)",
		"OAuth Owner Business", ownerID,
	)
	if err != nil {
		t.Fatalf("insert business: %v", err)
	}
	businessID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("business id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM businesses WHERE id = ?", businessID)
	})

	if _, err := db.ExecContext(ctx, "UPDATE users SET business_id = ? WHERE id = ?", businessID, ownerID); err != nil {
		t.Fatalf("set owner business: %v", err)
	}

	gotBusinessID, err := repo.GetUserBusinessID(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetUserBusinessID: %v", err)
	}
	if gotBusinessID != businessID {
		t.Fatalf("expected business=%d, got business=%d", businessID, gotBusinessID)
	}
}

func TestRevokeToken_NonExistentTokenIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Should not error even if the token does not exist.
	err := repo.RevokeToken(ctx, "non-existent-refresh-token")
	if err != nil {
		t.Fatalf("RevokeToken on non-existent token returned error: %v", err)
	}
}

func TestMarkAuthCodeUsed_NonExistentCodeIsNoOp(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkAuthCodeUsed(ctx, "non-existent-code")
	if err != nil {
		t.Fatalf("MarkAuthCodeUsed on non-existent code returned error: %v", err)
	}
}
