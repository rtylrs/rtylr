package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

const (
	insightsProjectPage = 200
	// insightsProjectCap bounds the projects one insights read covers.
	insightsProjectCap = 2000
)

// GetInsightsV2 is the business-wide Flow overview over the projects the
// caller sees in ListProjectsV2. Its task totals are those projects' progress
// counts, so the two screens agree.
func (s *Service) GetInsightsV2(ctx context.Context, callerID, businessID int64, now time.Time) (domain.Insights, error) {
	var projectIDs []int64
	var total, done int
	for offset := int64(0); offset < insightsProjectCap; offset += insightsProjectPage {
		page, count, err := s.ListProjectsV2(ctx, callerID, businessID, "", "", "", insightsProjectPage, offset)
		if err != nil {
			return domain.Insights{}, err
		}
		for _, p := range page {
			projectIDs = append(projectIDs, p.ID)
			total += int(p.TaskCount)
			done += int(p.CompletedCount)
		}
		if len(page) < insightsProjectPage || offset+insightsProjectPage >= count {
			break
		}
	}

	ins, err := s.repo.GetInsightsV2(ctx, businessID, projectIDs, now.UTC())
	if err != nil {
		return domain.Insights{}, err
	}
	ins.TotalTasks, ins.DoneTasks = total, done
	if len(projectIDs) == 0 {
		return ins, nil
	}
	ins.TimeByMember, err = s.insightsTimeByMember(ctx, businessID, ins.TimeByMember,
		func(afterID int64, limit int32) ([]int64, error) {
			return s.repo.ListTaskIDsInProjects(ctx, businessID, projectIDs, afterID, limit)
		})
	if err != nil {
		return domain.Insights{}, err
	}
	return ins, nil
}
