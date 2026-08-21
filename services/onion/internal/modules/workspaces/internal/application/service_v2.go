package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// DefaultBinRetentionDays is the recycle-bin window used until the per-business
// retention config key lands (spec ruling P2/D-decision: configurable, 30d
// default). Archive stamps purge_at from this at archive time; later config
// changes affect future archives only.
const DefaultBinRetentionDays = 30

// ErrNotFound is returned by v2 lookups when the business has no such project.
var ErrNotFound = errors.New("flow project not found")

// canSeeProject reports whether callerID may READ projectID under the curation
// rule: creator, member, or hub bypass. Public visibility never widens read
// access here - public projects surface as metadata on profiles, not as
// openable boards.
func (s *Service) canSeeProject(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if s.hubProfileBypasses(ctx, callerID, businessID) {
		return true
	}
	if p.CreatedBy == callerID || p.OwnerID == callerID {
		return true
	}
	return s.repo.IsProjectMember(ctx, callerID, p.ID)
}

// GetProjectV2 returns one project if the caller may see it.
func (s *Service) GetProjectV2(ctx context.Context, projectID, callerID, businessID int64) (domain.Project, error) {
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return domain.Project{}, err
	}
	if !s.canSeeProject(ctx, p, callerID, businessID) {
		return domain.Project{}, domain.ErrUnauthorized
	}
	return p, nil
}

// ListProjectsV2 lists active non-personal projects under curation. Bypass
// callers see the whole business; everyone else sees created + member.
func (s *Service) ListProjectsV2(ctx context.Context, callerID, businessID int64, status, visibility, q string, limit, offset int64) ([]domain.Project, int64, error) {
	includeAll := s.hubProfileBypasses(ctx, callerID, businessID)
	return s.repo.ListProjectsV2(ctx, businessID, callerID, includeAll, status, visibility, q, limit, offset)
}

// ListArchivedProjectsV2 lists the caller's visible recycle-bin entries,
// soonest purge first.
func (s *Service) ListArchivedProjectsV2(ctx context.Context, callerID, businessID int64, limit, offset int64) ([]domain.Project, int64, error) {
	includeAll := s.hubProfileBypasses(ctx, callerID, businessID)
	return s.repo.ListArchivedProjectsV2(ctx, businessID, callerID, includeAll, limit, offset)
}

// SetProjectVisibility flips private/public. Gate mirrors other project admin
// operations; visibility changes invalidate the workspaces scope because list
// projections may embed visibility-derived ordering later.
func (s *Service) SetProjectVisibility(ctx context.Context, projectID, callerID, businessID int64, visibility string) (domain.Project, error) {
	if !domain.ValidVisibility(visibility) {
		return domain.Project{}, errors.New("invalid visibility")
	}
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Project{}, err
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return domain.Project{}, err
	}
	if p.IsPersonal {
		return domain.Project{}, errors.New("the My Tasks bucket is always private")
	}
	if err := s.repo.UpdateProjectVisibility(ctx, projectID, businessID, visibility); err != nil {
		return domain.Project{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.GetProjectV2ByID(ctx, projectID, businessID)
}

// RestoreProjectV2 pulls a binned project back to active. Admin gate matches
// archive; the backfilled membership model guarantees former workspace admins
// hold this role where they had it before.
func (s *Service) RestoreProjectV2(ctx context.Context, projectID, callerID, businessID int64) (domain.Project, error) {
	if err := s.RequireProjectRole(ctx, projectID, callerID, businessID, domain.FlowRoleAdmin); err != nil {
		return domain.Project{}, err
	}
	if err := s.repo.RestoreProjectV2(ctx, projectID, businessID); err != nil {
		return domain.Project{}, err
	}
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.GetProjectV2ByID(ctx, projectID, businessID)
}

// PurgeProjectV2 irreversibly deletes an archived project. Gate is deliberately
// narrower than restore: the creator or a hub owner/admin only (ruling P4).
// Returns the attachment object URLs that were referenced so the caller can
// sweep stored files once rows are gone.
func (s *Service) PurgeProjectV2(ctx context.Context, projectID, callerID, businessID int64) ([]string, error) {
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return nil, err
	}
	if p.ArchivedAt == nil {
		return nil, errors.New("only archived projects can be purged")
	}
	bypass := s.hubProfileBypasses(ctx, callerID, businessID)
	if !bypass && p.CreatedBy != callerID && p.OwnerID != callerID {
		return nil, domain.ErrUnauthorized
	}
	urls, err := s.repo.PurgeProjectCascade(ctx, projectID, businessID)
	if err != nil {
		return nil, err
	}
	invalidateWorkspacesScope(ctx, businessID)
	return urls, nil
}

// binPurgeDeadline computes the permanent-deletion timestamp for an archive
// happening now.
func binPurgeDeadline(now time.Time) time.Time {
	return now.AddDate(0, 0, DefaultBinRetentionDays)
}

// IsProjectMemberOrBypass is the v2 read gate for project-scoped subresources:
// hub bypass, creator, owner, or explicit membership.
func (s *Service) IsProjectMemberOrBypass(ctx context.Context, userID, businessID, projectID int64) bool {
	if s.hubProfileBypasses(ctx, userID, businessID) {
		return true
	}
	p, err := s.repo.GetProjectV2ByID(ctx, projectID, businessID)
	if err != nil {
		return false
	}
	if p.CreatedBy == userID || p.OwnerID == userID {
		return true
	}
	return s.repo.IsProjectMember(ctx, userID, p.ID)
}
