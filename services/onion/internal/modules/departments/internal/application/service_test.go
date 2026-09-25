package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/departments/internal/domain"
)

type mockRepo struct {
	createID  int64
	createErr error
	updErr    error
	delErr    error
	list      []domain.Department
}

func (m *mockRepo) ListDepartments(_ context.Context, _ int64) ([]domain.Department, error) {
	return m.list, nil
}

func (m *mockRepo) CreateDepartment(_ context.Context, _ domain.DepartmentInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) UpdateDepartment(_ context.Context, _ int64, _ domain.DepartmentInput) error {
	return m.updErr
}

func (m *mockRepo) DeleteDepartment(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Department{{ID: 1, Name: "Eng"}}}}
	got, _ := s.ListDepartments(context.Background(), 1)
	if len(got) != 1 || got[0].Name != "Eng" {
		t.Errorf("got %v", got)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateDepartment(context.Background(), domain.DepartmentInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateDepartment(context.Background(), 1, domain.DepartmentInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.DeleteDepartment(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
