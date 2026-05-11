package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
)

type mockRepo struct {
	steps    []*domain.SetupStep
	cfg      *domain.HRConfig
	stepsErr error
	upStepEr error
	cfgErr   error
	upCfgErr error
}

func (m *mockRepo) GetSetupSteps(_ context.Context, _ int64) ([]*domain.SetupStep, error) {
	return m.steps, m.stepsErr
}

func (m *mockRepo) UpsertSetupStep(_ context.Context, _ int64, _ string, _ time.Time) error {
	return m.upStepEr
}

func (m *mockRepo) GetHRConfig(_ context.Context, _ int64) (*domain.HRConfig, error) {
	return m.cfg, m.cfgErr
}

func (m *mockRepo) UpsertHRConfig(_ context.Context, _ int64, _ *time.Time) (*domain.HRConfig, error) {
	return m.cfg, m.upCfgErr
}

func TestGetSetupStatus_AllStepsComplete(t *testing.T) {
	now := time.Now()
	steps := []*domain.SetupStep{}
	for k := range validStepKeys {
		steps = append(steps, &domain.SetupStep{StepKey: k, CompletedAt: &now})
	}
	s := &Service{repo: &mockRepo{steps: steps}}
	_, complete, err := s.GetSetupStatus(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !complete {
		t.Error("expected complete=true when all steps done")
	}
}

func TestGetSetupStatus_PartiallyComplete(t *testing.T) {
	now := time.Now()
	steps := []*domain.SetupStep{
		{StepKey: "company_profile", CompletedAt: &now},
		{StepKey: "payroll", CompletedAt: nil},
	}
	s := &Service{repo: &mockRepo{steps: steps}}
	_, complete, _ := s.GetSetupStatus(context.Background(), 1)
	if complete {
		t.Error("expected complete=false")
	}
}

func TestUpsertHRConfig_PassesNilTimestamp(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.UpsertHRConfig(context.Background(), 1, nil); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestUpsertHRConfig_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{upCfgErr: want}}
	if _, err := s.UpsertHRConfig(context.Background(), 1, nil); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCompleteSetupStep_RejectsUnknownKey(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	for _, key := range []string{"", "fake_step", "company_PROFILE", "Payroll"} {
		err := s.CompleteSetupStep(context.Background(), 1, key)
		if !errors.Is(err, domain.ErrInvalidStepKey) {
			t.Errorf("expected ErrInvalidStepKey for %q, got %v", key, err)
		}
	}
}

func TestCompleteSetupStep_HappyPathDelegatesToRepo(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.CompleteSetupStep(context.Background(), 1, "payroll"); err != nil {
		t.Errorf("unexpected: %v", err)
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
