package application

import (
	"context"
	"sort"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
)

// ProjectAccess is Flow's own answer to who may see a project's reporting,
// served by the workspaces module. The project dashboard asks it rather than
// restating Flow's privacy rules in SQL here.
type ProjectAccess interface {
	// CanSeeProjectReport is true for anyone Flow lets open the project and
	// for anyone with Flow's business-wide reporting visibility.
	CanSeeProjectReport(ctx context.Context, businessID, userID, projectID int64) (bool, error)
	// ListedProjectIDs is what Flow's sidebar lists for userID (ruling P8).
	ListedProjectIDs(ctx context.Context, businessID, userID int64) ([]int64, error)
}

// SetProjectAccess wires Flow's project access decision. Until it is called the
// dashboard falls back to its own rule alone: public projects, and private ones
// the viewer owns, created or is a member of.
func (s *Service) SetProjectAccess(access ProjectAccess) {
	s.access = access
}

// ListDashboardProjects returns the projects userID may pick a dashboard for:
// every public project and the private ones they own, created or belong to,
// plus whatever else Flow's sidebar lists for them, such as their reports'
// projects. A bypass role's picker stays curated the way Flow's sidebar is,
// even though it may open any project's dashboard by id.
func (s *Service) ListDashboardProjects(ctx context.Context, businessID, userID int64) ([]domain.DashboardProject, error) {
	projects, err := s.repo.ListDashboardProjects(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	if s.access == nil {
		return projects, nil
	}
	listed, err := s.access.ListedProjectIDs(ctx, businessID, userID)
	if err != nil {
		return nil, err
	}
	seen := make(map[int64]bool, len(projects))
	for _, p := range projects {
		seen[p.ID] = true
	}
	missing := make([]int64, 0, len(listed))
	for _, id := range listed {
		if !seen[id] {
			seen[id] = true
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return projects, nil
	}
	extra, err := s.repo.ListDashboardProjectsByIDs(ctx, businessID, missing)
	if err != nil {
		return nil, err
	}
	projects = append(projects, extra...)
	sort.SliceStable(projects, func(i, j int) bool {
		a, b := strings.ToLower(projects[i].Name), strings.ToLower(projects[j].Name)
		if a != b {
			return a < b
		}
		return projects[i].ID < projects[j].ID
	})
	return projects, nil
}

// GetDashboardProject returns a project userID may open a dashboard for, or
// found false. The dashboard's own rule settles the common case in one query;
// anything it does not cover is put to Flow, which admits its bypass roles and
// a manager whose report is involved in the project.
func (s *Service) GetDashboardProject(ctx context.Context, businessID, userID, projectID int64) (domain.DashboardProject, bool, error) {
	project, found, err := s.repo.GetDashboardProject(ctx, businessID, userID, projectID)
	if err != nil || found || s.access == nil {
		return project, found, err
	}
	allowed, err := s.access.CanSeeProjectReport(ctx, businessID, userID, projectID)
	if err != nil || !allowed {
		return domain.DashboardProject{}, false, err
	}
	return s.repo.GetDashboardProjectByID(ctx, businessID, projectID)
}
