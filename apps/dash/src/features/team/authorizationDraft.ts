import type { AppRoleAssignment } from '@voxire/voxie/hub/user/v1/types';

export function legacyRoleToDashRole(role: string): string {
  switch (role.toLowerCase()) {
    case 'owner':
      return 'owner';
    case 'admin':
      return 'admin';
    case 'operator':
    case 'billing':
    case 'menu_manager':
    case 'store_manager':
      return 'manager';
    case 'viewer':
      return 'viewer';
    default:
      return 'member';
  }
}

/**
 * Keeps the legacy API enum populated during migration without making it the
 * role shown or edited by Dash.
 */
export function dashRoleToLegacyRole(role: string): string {
  switch (role) {
    case 'admin':
      return 'admin';
    case 'manager':
      return 'operator';
    case 'viewer':
      return 'viewer';
    case 'owner':
      return 'admin';
    case 'member':
    default:
      return 'staff';
  }
}

export function buildAppRoleAssignments(
  apps: string[],
  groupId?: number,
  roleIdsByApp: Record<string, string[]> = {}
): AppRoleAssignment[] {
  const source = groupId === undefined ? 'explicit' : 'group';
  return apps.map((appId) => {
    const roleIds = roleIdsByApp[appId] ?? [];
    return {
      appId,
      roleIds,
      ...(roleIds.length > 0 ? { primaryRoleId: roleIds[0] } : {}),
      assignmentSource: source,
      status: 'active',
      scopes: [],
      ...(groupId === undefined ? {} : { groupId: String(groupId) }),
    };
  });
}
