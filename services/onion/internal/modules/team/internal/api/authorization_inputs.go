package api

import (
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

func appAssignmentsFromProto(values []*hubuserv1.AppRoleAssignment) ([]domain.AppRoleAssignmentInput, error) {
	assignments := make([]domain.AppRoleAssignmentInput, 0, len(values))
	for _, value := range values {
		if value == nil || value.GetAppId() == "" {
			return nil, fmt.Errorf("app assignment requires an app id")
		}
		roleIDs := make([]int64, 0, len(value.GetRoleIds()))
		for _, rawRoleID := range value.GetRoleIds() {
			roleID, err := strconv.ParseInt(rawRoleID, 10, 64)
			if err != nil || roleID <= 0 {
				return nil, fmt.Errorf("invalid role id: %s", rawRoleID)
			}
			roleIDs = append(roleIDs, roleID)
		}

		var primaryRoleID *int64
		if value.PrimaryRoleId != nil {
			parsed, err := strconv.ParseInt(value.GetPrimaryRoleId(), 10, 64)
			if err != nil || parsed <= 0 {
				return nil, fmt.Errorf("invalid primary role id: %s", value.GetPrimaryRoleId())
			}
			primaryRoleID = &parsed
		}

		var groupID *int64
		if value.GroupId != nil {
			parsed, err := strconv.ParseInt(value.GetGroupId(), 10, 64)
			if err != nil || parsed <= 0 {
				return nil, fmt.Errorf("invalid group id: %s", value.GetGroupId())
			}
			groupID = &parsed
		}

		scopes := make([]authorization.Scope, 0, len(value.GetScopes()))
		for _, scope := range value.GetScopes() {
			if scope == nil || scope.GetScopeType() == "" || scope.GetScopeId() == "" {
				return nil, fmt.Errorf("role assignment scope requires a type and id")
			}
			scopes = append(scopes, authorization.Scope{Type: scope.GetScopeType(), ID: scope.GetScopeId()})
		}

		assignments = append(assignments, domain.AppRoleAssignmentInput{
			AppID:         value.GetAppId(),
			RoleIDs:       roleIDs,
			PrimaryRoleID: primaryRoleID,
			Source:        value.GetAssignmentSource(),
			Status:        value.GetStatus(),
			Scopes:        scopes,
			GroupID:       groupID,
		})
	}
	return assignments, nil
}
