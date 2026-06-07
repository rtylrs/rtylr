package http

import (
	"testing"

	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
)

func TestValidateProtoMessagesRejectsInvalidGeneratedMessage(t *testing.T) {
	req := &commonauthv1.LoginRequest{}

	if err := validateProtoMessages(req); err == nil {
		t.Fatal("expected proto validation error")
	}
}

func TestValidateProtoMessagesAcceptsNestedGeneratedMessage(t *testing.T) {
	res := struct {
		Payload *commonauthv1.LoginRequest
	}{
		Payload: &commonauthv1.LoginRequest{
			Email:    "person@example.com",
			Password: "secret",
		},
	}

	if err := validateProtoMessages(res); err != nil {
		t.Fatalf("validateProtoMessages: %v", err)
	}
}
