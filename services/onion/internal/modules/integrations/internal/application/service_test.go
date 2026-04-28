package application

import (
	"context"
	"testing"
)

func TestConnect_RequiresProvider(t *testing.T) {
	s := &Service{}
	if _, err := s.Connect(context.Background(), 1, "", nil); err == nil {
		t.Fatalf("expected error for empty provider")
	}
}

func TestDisconnect_RequiresIntegrationID(t *testing.T) {
	s := &Service{}
	if err := s.Disconnect(context.Background(), 1, 0); err == nil {
		t.Fatalf("expected error for missing integration id")
	}
}
