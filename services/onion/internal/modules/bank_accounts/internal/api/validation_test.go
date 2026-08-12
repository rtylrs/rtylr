package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	bankaccountv1 "github.com/voxire/voxie/gen/go/finance/bank_account/v1"
)

func TestCreateBankAccountRejectsInvalidInputBeforeService(t *testing.T) {
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{BusinessID: 1, UserID: 1})
	_, err := (&BankAccountAPI{}).CreateBankAccount(ctx, &bankaccountv1.CreateBankAccountRequest{Balance: -1})

	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an API error, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode() != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", apiErr.HTTPStatusCode())
	}
}
