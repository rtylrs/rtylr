package voxie_test

import (
	"testing"

	posuserv1 "github.com/voxire/voxie/gen/go/pos/user/v1"
)

func TestUpdatePOSUserAllowsOmittedPIN(t *testing.T) {
	request := &posuserv1.UpdatePOSUserRequest{
		Id:             1,
		Name:           "Cashier",
		Role:           "cashier",
		MaxDiscountPct: 0,
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("expected an update without a replacement PIN to validate: %v", err)
	}
}

func TestCreatePOSUserStillRequiresPIN(t *testing.T) {
	request := &posuserv1.CreatePOSUserRequest{
		LocationId:     1,
		Name:           "Cashier",
		Role:           "cashier",
		MaxDiscountPct: 0,
	}

	if err := request.Validate(); err == nil {
		t.Fatal("expected create without a PIN to fail validation")
	}
}
