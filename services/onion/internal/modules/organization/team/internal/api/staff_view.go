package api

import (
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/domain"
	"github.com/voxire/onion/internal/platform/authorization"
)

type staffPayload struct {
	ID                           int64
	MemberType                   string
	Email                        string
	FullName                     string
	CreatedAt                    time.Time
	LicensedAppIDs               []string
	PermissionProfileAssignments []authorization.RoleAssignment
	EffectiveAccess              *authorization.EffectiveAccess
	IsActive                     bool
	Groups                       []domain.GroupSummary
}

func staffPayloads(values []domain.StaffUser) []staffPayload {
	if values == nil {
		return nil
	}
	out := make([]staffPayload, len(values))
	for i, v := range values {
		out[i] = staffPayload(v)
	}
	return out
}
