package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow_insights/internal/domain"
)

type mockRepo struct {
	stats   domain.InsightsStats
	team    domain.TeamWorkloadResult
	roadmap domain.RoadmapResult
	err     error
}

func (m *mockRepo) GetInsightsStats(_ context.Context, _ int64) (domain.InsightsStats, error) {
	return m.stats, m.err
}

func (m *mockRepo) GetTeamWorkload(_ context.Context, _ int64) (domain.TeamWorkloadResult, error) {
	return m.team, m.err
}

func (m *mockRepo) GetRoadmap(_ context.Context, _ int64) (domain.RoadmapResult, error) {
	return m.roadmap, m.err
}

func TestGetInsights_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.GetInsights(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetTeamWorkload_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetTeamWorkload(context.Background(), 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestGetRoadmap_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetRoadmap(context.Background(), 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
