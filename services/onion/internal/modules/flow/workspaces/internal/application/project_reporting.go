package application

import (
	"context"
	"database/sql"
	"errors"
)

// listedProjectsPage and listedProjectsCap bound ListedProjectIDs. The cap
// matches the project dashboard picker's own limit of 500 projects.
const (
	listedProjectsPage int64 = 100
	listedProjectsCap  int64 = 500
)

// CanSeeProjectReport reports whether callerID may see reporting on one
// project, such as its project dashboard. That is exactly who Flow lets open
// the project (canSeeProject: creator, owner, member, a manager of someone
// involved in it, or the business owner, an admin or a Flow Admin), so Insights
// and Work agree on who sees a project.
//
// The My Tasks bucket is never reported on, whoever asks, which is the one
// difference from opening it: its owner opens their own bucket but gets no
// dashboard for it. A project that does not exist or belongs to another
// business is false with no error, so a caller cannot tell it apart from one it
// may not see.
func (s *Service) CanSeeProjectReport(ctx context.Context, projectID, callerID, businessID int64) (bool, error) {
	if projectID <= 0 || callerID <= 0 || businessID <= 0 {
		return false, nil
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if p.IsPersonal {
		return false, nil
	}
	return s.canSeeProject(ctx, p, callerID, businessID), nil
}

// ListedProjectIDs returns the ids of the projects Flow's sidebar lists for
// callerID, which is ListProjectsV2 under ruling P8: what they created or
// belong to plus their reports' projects, for bypass roles too. It reads every
// page up to listedProjectsCap.
func (s *Service) ListedProjectIDs(ctx context.Context, callerID, businessID int64) ([]int64, error) {
	ids := []int64{}
	if callerID <= 0 || businessID <= 0 {
		return ids, nil
	}
	for offset := int64(0); offset < listedProjectsCap; offset += listedProjectsPage {
		projects, total, err := s.ListProjectsV2(ctx, callerID, businessID, "", "", "", listedProjectsPage, offset)
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			ids = append(ids, p.ID)
		}
		if int64(len(projects)) < listedProjectsPage || offset+listedProjectsPage >= total {
			break
		}
	}
	return ids, nil
}
