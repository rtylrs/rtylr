package domain

// Flow's per-project and per-workspace member roles. These are the values
// stored in flow_project_members.role and flow_workspace_members.role.
const (
	FlowRoleViewer = "viewer"
	FlowRoleMember = "member"
	FlowRoleAdmin  = "admin"

	// FlowRoleManager is a legacy API alias for the top tier. The column is
	// ENUM('admin','member','viewer') and cannot store it, so it is treated
	// as an alias of FlowRoleAdmin rather than a rank of its own.
	FlowRoleManager = "manager"
)

// CanonicalFlowMemberRole accepts the public API's legacy manager alias and
// returns the value that can be stored in the Flow role enum.
func CanonicalFlowMemberRole(role string) (string, bool) {
	switch role {
	case FlowRoleViewer, FlowRoleMember, FlowRoleAdmin:
		return role, true
	case FlowRoleManager:
		return FlowRoleAdmin, true
	default:
		return "", false
	}
}

// Role tiers, ordered viewer < member < admin. Zero means "not a role we
// recognise" and never satisfies anything.
const (
	rankUnknown = 0
	rankViewer  = 1
	rankMember  = 2
	rankAdmin   = 3
)

// RankFlowRole maps a stored role to its tier. Anything unrecognised ranks
// zero so callers fail closed rather than admitting an unexpected value.
func RankFlowRole(role string) int {
	switch role {
	case FlowRoleViewer:
		return rankViewer
	case FlowRoleMember:
		return rankMember
	case FlowRoleAdmin, FlowRoleManager:
		return rankAdmin
	default:
		return rankUnknown
	}
}

// RoleAllows reports whether a caller holding role satisfies minRole. An
// unrecognised value on either side denies: an unknown role is not trusted,
// and an unknown requirement is not satisfiable.
func RoleAllows(role, minRole string) bool {
	held, required := RankFlowRole(role), RankFlowRole(minRole)
	if held == rankUnknown || required == rankUnknown {
		return false
	}
	return held >= required
}
