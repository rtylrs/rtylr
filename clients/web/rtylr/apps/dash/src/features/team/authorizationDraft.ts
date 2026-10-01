import type { PermissionProfileAssignment } from '../../shared/schemas/team';

export function buildPermissionProfileAssignments(
  appIds: string[],
  profileIdsByApp: Record<string, string[]>,
  assignmentSource: 'explicit' | 'group' = 'explicit'
): PermissionProfileAssignment[] {
  return appIds.flatMap((appId) => {
    const permissionProfileIds = [...new Set(profileIdsByApp[appId] ?? [])];
    if (permissionProfileIds.length === 0) return [];
    return [
      {
        appId,
        permissionProfileIds,
        assignmentSource,
        status: 'active',
        scopes: [],
      },
    ];
  });
}
