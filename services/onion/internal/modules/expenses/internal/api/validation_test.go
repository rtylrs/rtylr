package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	expensesv1 "github.com/voxire/voxie/gen/go/finance/expenses/v1"
)

func TestCreateExpenseRejectsInvalidInputBeforeService(t *testing.T) {
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{BusinessID: 1, UserID: 1})
	_, err := (&ExpensesAPI{}).CreateExpense(ctx, &expensesv1.CreateExpenseRequest{
		Amount:      1,
		ExpenseDate: "2026-01-01",
	})

	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an API error, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode() != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing description, got %d", apiErr.HTTPStatusCode())
	}
}

func TestUpdateExpenseRejectsInvalidInputBeforeService(t *testing.T) {
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{BusinessID: 1, UserID: 1})
	_, err := (&ExpensesAPI{}).UpdateExpense(ctx, &expensesv1.UpdateExpenseRequest{
		Id: 1,
		Body: &expensesv1.UpdateExpenseBody{
			Amount:      1,
			ExpenseDate: "2026-01-01",
		},
	})

	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an API error, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode() != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing description, got %d", apiErr.HTTPStatusCode())
	}
}
