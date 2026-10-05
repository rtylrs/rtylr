package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// DefaultBinRetentionDays is the recycle-bin window used until the per-business
// retention config key lands (spec ruling P2/D-decision: configurable, 30d
// default). Archive stamps purge_at from this at archive time; later config
// changes affect future archives only.
const DefaultBinRetentionDays = 30

// ErrNotFound is returned by v2 lookups when the business has no such project.
var ErrNotFound = errors.New("flow project not found")

// ErrProjectNotArchived is a client error, not a fault: purging is only
// defined for an archived project, and a caller who has not archived one yet
// needs to be told that rather than shown a 500.
var ErrProjectNotArchived = errors.New("only archived projects can be purged")

// canSeeProject reports whether callerID may READ projectID under the curation
// rule: creator, member, or manager visibility through the reporting line.
// Public visibility never widens read access here.
func (s *Service) canSeeProject(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.CreatedBy == callerID || p.OwnerID == callerID {
		return true
	}
	if s.repo.IsProjectMember(ctx, callerID, p.ID) {
		return true
	}
	return s.managesSomeoneOn(ctx, p, callerID, businessID)
}

// managesSomeoneOn reports whether the caller manages anyone involved in this
// project, which is the dept-manager visibility proxy (M3).
//
// The My Tasks bucket is exempt no matter who manages its owner: Project's own
// contract says is_personal is "exempt from manager-visibility policies unless
// its owner explicitly shares it", so the check stops before the proxy.
//
// Involvement means created, owns, or is a member. Membership is probed one
// report at a time rather than with an IN clause because a manager has few
// reports and sqlc's slice expansion produces invalid SQL for an empty set in
// MySQL. It does mean this is O(reports) queries on a cache miss, so it runs
// last, after the cheap self checks have already returned.
func (s *Service) managesSomeoneOn(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.IsPersonal {
		return false
	}
	reports := s.managedUserIDs(ctx, callerID, businessID)

	// Creator and owner are already in hand, so settle the common case without
	// touching the database at all.
	for _, reportID := range reports {
		if p.CreatedBy == reportID || p.OwnerID == reportID {
			return true
		}
	}

	// Only then pay for membership lookups. A nil repo means the proxy cannot
	// answer, and the rule throughout is that failing to resolve must decline to
	// widen rather than widen blindly.
	if s.repo == nil {
		return false
	}
	for _, reportID := range reports {
		if s.repo.IsProjectMember(ctx, reportID, p.ID) {
			return true
		}
	}
	return false
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
	// Ruling P8: this listing is curated for EVERYONE, bypass roles included.
	// A central Flow manager holds business-wide management permission, but
	// access is not listing: their sidebar shows what they created or belong to, like anyone
	// else's. Passing the bypass through here put every private project in the
	// business into an admin's sidebar, which is the precise thing P8 forbids.
	//
	// Deliberately not applied to ListArchivedProjectsV2: the recycle bin is
	// governed by P4, which gives the business owner and admins restore rights
	// and therefore sight of the entries.
	const includeAll = false

	// Resolved unconditionally now. Previously skipped for bypass callers on the
	// grounds that they already saw everything; with the listing curated, a
	// manager who also holds a bypass role still needs their reports' projects
	// grouped under P9.
	managed := s.managedUserIDs(ctx, callerID, businessID)

	projects, total, err := s.repo.ListProjectsV2(ctx, businessID, callerID, includeAll, managed, status, visibility, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return s.markVisibleViaReport(ctx, projects, callerID, managed), total, nil
}

// markVisibleViaReport flags the projects this caller sees only through the
// dept-manager proxy, so a client can group them (P9) without asking per project
// why each one is visible.
//
// A project the caller created, owns, or is a member of is theirs, so the flag
// stays false. Anything else in the result must have matched a managed branch,
// because that is the only other way the list query admits a row. With no
// reports the loop does not run, which is the common case and costs nothing.
func (s *Service) markVisibleViaReport(ctx context.Context, projects []domain.Project, callerID int64, managed []int64) []domain.Project {
	if len(managed) == 0 || len(projects) == 0 {
		return projects
	}
	for i := range projects {
		p := &projects[i]
		// Creator and owner are already in hand, so a caller's own project is
		// settled without a query. This check comes first deliberately: it is
		// the one that must hold even when the membership lookup cannot run.
		if p.CreatedBy == callerID || p.OwnerID == callerID {
			continue
		}
		// Without a repo the membership question cannot be answered. Flagging
		// optimistically would advertise a caller's own project as a report's,
		// so leave it false and degrade to no grouping rather than wrong
		// grouping.
		if s.repo == nil {
			continue
		}
		if s.repo.IsProjectMember(ctx, callerID, p.ID) {
			continue
		}
		p.VisibleViaReport = true
	}
	return projects
}

// ListArchivedProjectsV2 lists the caller's visible recycle-bin entries,
// soonest purge first.
func (s *Service) ListArchivedProjectsV2(ctx context.Context, callerID, businessID, limit, offset int64) ([]domain.Project, int64, error) {
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
		return nil, ErrProjectNotArchived
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
