package api

import (
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

func permissionProfileAssignmentsFromProto(values []*hubuserv1.PermissionProfileAssignment) ([]domain.AppRoleAssignmentInput, error) {
	assignments := make([]domain.AppRoleAssignmentInput, 0, len(values))
	for _, value := range values {
		if value == nil || value.GetAppId() == "" {
			return nil, fmt.Errorf("permission profile assignment requires an app id")
		}
		roleIDs := make([]int64, 0, len(value.GetPermissionProfileIds()))
		for _, rawRoleID := range value.GetPermissionProfileIds() {
			roleID, err := strconv.ParseInt(rawRoleID, 10, 64)
			if err != nil || roleID <= 0 {
				return nil, fmt.Errorf("invalid permission profile id: %s", rawRoleID)
			}
			roleIDs = append(roleIDs, roleID)
		}

		scopes := make([]authorization.Scope, 0, len(value.GetScopes()))
		for _, scope := range value.GetScopes() {
			if scope == nil || scope.GetScopeType() == "" || scope.GetScopeId() == "" {
				return nil, fmt.Errorf("permission scope requires a type and id")
			}
			scopes = append(scopes, authorization.Scope{Type: scope.GetScopeType(), ID: scope.GetScopeId()})
		}

		assignments = append(assignments, domain.AppRoleAssignmentInput{
			AppID:   value.GetAppId(),
			RoleIDs: roleIDs,
			Source:  string(authorization.SourceExplicit),
			Status:  "active",
			Scopes:  scopes,
		})
	}
	return assignments, nil
}
