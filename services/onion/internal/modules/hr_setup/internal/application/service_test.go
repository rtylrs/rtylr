package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
)

func TestCompleteSetupStep_RejectsUnknownKey(t *testing.T) {
	s := &Service{}
	for _, key := range []string{"", "fake_step", "company_PROFILE", "Payroll"} {
		err := s.CompleteSetupStep(context.Background(), 1, key)
		if !errors.Is(err, domain.ErrInvalidStepKey) {
			t.Errorf("expected ErrInvalidStepKey for %q, got %v", key, err)
		}
	}
}

func TestValidStepKeys_CoversExpectedSurface(t *testing.T) {
	expected := []string{
		"company_profile",
		"work_schedule",
		"org_structure",
		"locations",
		"roles_permissions",
		"payroll",
		"notifications",
		"employees",
	}
	for _, k := range expected {
		if _, ok := validStepKeys[k]; !ok {
			t.Errorf("expected %q to be a valid step key", k)
		}
	}
	if len(validStepKeys) != len(expected) {
		t.Errorf("validStepKeys has %d entries, expected %d (changing the set is a contract change)",
			len(validStepKeys), len(expected))
	}
}
