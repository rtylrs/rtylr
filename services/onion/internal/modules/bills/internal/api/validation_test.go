package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	billv1 "github.com/voxire/voxie/gen/go/finance/bill/v1"
)

func TestCreateBillRejectsInvalidInputBeforeService(t *testing.T) {
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{BusinessID: 1, UserID: 1})
	_, err := (&BillAPI{}).CreateBill(ctx, &billv1.CreateBillRequest{DueDate: "2026-01-01"})

	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an API error, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode() != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-positive amount, got %d", apiErr.HTTPStatusCode())
	}
}

func TestUpdateBillRejectsInvalidInputBeforeService(t *testing.T) {
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{BusinessID: 1, UserID: 1})
	_, err := (&BillAPI{}).UpdateBill(ctx, &billv1.UpdateBillRequest{
		Id: 1,
		Body: &billv1.UpdateBillBody{
			Amount:  1,
			DueDate: "2026-01-01",
			Status:  "not-a-status",
		},
	})

	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an API error, got %T: %v", err, err)
	}
	if apiErr.HTTPStatusCode() != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", apiErr.HTTPStatusCode())
	}
}
