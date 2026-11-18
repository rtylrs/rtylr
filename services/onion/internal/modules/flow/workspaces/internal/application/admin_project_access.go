package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// administratorOpens reports whether callerID may open p because they own or
// administer the business it belongs to.
//
// This is access, not listing (ruling P8): the sidebar stays curated, and an
// administrator reaches a private project by link, the Insights project picker
// or the recycle bin. It mirrors the business owner/admin pass that
// RequireProjectRole already gives writes, so an administrator who may edit a
// project may also open it.
//
// Two things stay closed. A My Tasks bucket belongs to one person and is exempt
// from every visibility policy, so no administrator opens someone else's. And
// the project must belong to the business the caller is acting in, so being an
// administrator somewhere never opens a project somewhere else.
func (s *Service) administratorOpens(ctx context.Context, p domain.Project, callerID, businessID int64) bool {
	if p.IsPersonal || p.BusinessID != businessID {
		return false
	}
	return s.isBusinessAdministrator(ctx, callerID, businessID)
}
