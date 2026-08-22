package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	commonpaginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
)

const (
	defaultPageLimit = int32(50)
	maxPageLimit     = int32(500)
)

// resolvePagination applies this repo's onekit pagination defaults (default
// 50, cap 500) to a request's page/limit fields and derives the SQL offset -
// the Flow contracts carry page/limit, not offset directly, so offset is
// computed the same way the rest of the onekit path derives page from offset:
// offset = (page - 1) * limit.
func resolvePagination(page, limit int32) (resolvedPage, resolvedLimit int32, offset int64) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	return page, limit, int64(page-1) * int64(limit)
}

// buildPagination assembles the common.pagination.v1 response envelope from a
// resolved page/limit, the real row count from a COUNT query, and the number
// of rows actually returned for this page.
func buildPagination(page, limit int32, total int64, returned int) *commonpaginationv1.PaginationResponse {
	var totalPages int32
	if limit > 0 {
		totalPages = int32((total + int64(limit) - 1) / int64(limit))
	}
	hasNext := int64(page-1)*int64(limit)+int64(returned) < total
	return &commonpaginationv1.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      int32(total),
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    page > 1,
	}
}

func resolveWorkspaceQuery(ctx context.Context, svc *application.Service, authUser *authPkg.AuthUser, businessID int64, workspaceID string) (int64, error) {
	wsID := businessID
	if workspaceID == "" {
		if !svc.IsWorkspaceMember(ctx, authUser.UserID, wsID) {
			return 0, onekiterrors.PermissionDenied()
		}
		return wsID, nil
	}
	resolved, err := svc.ResolveWorkspaceID(ctx, authUser.UserID, businessID, workspaceID)
	if err != nil || resolved == 0 {
		return 0, onekiterrors.PermissionDenied()
	}
	return resolved, nil
}
