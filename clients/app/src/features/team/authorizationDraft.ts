import type { PermissionProfileAssignment } from '../../shared/schemas/team';

export function getAssignedProfileIdsByApp(
  assignments: PermissionProfileAssignment[],
  assignmentSource: 'explicit' | 'group'
): Record<string, string[]> {
  const byApp: Record<string, string[]> = {};
  for (const assignment of assignments) {
    if (assignment.assignmentSource !== assignmentSource || assignment.status !== 'active')
      continue;
    byApp[assignment.appId] = [
      ...new Set([...(byApp[assignment.appId] ?? []), ...assignment.permissionProfileIds]),
    ];
  }
  return byApp;
}

export function buildPermissionProfileAssignments(
  appIds: string[],
  profileIdsByApp: Record<string, string[]>,
  assignmentSource: 'explicit' | 'group' = 'explicit',
  existingAssignments: PermissionProfileAssignment[] = []
): PermissionProfileAssignment[] {
  return appIds.flatMap((appId) => {
    const permissionProfileIds = [...new Set(profileIdsByApp[appId] ?? [])];
    if (permissionProfileIds.length === 0) return [];
    const retained = existingAssignments
      .filter(
        (assignment) =>
          assignment.appId === appId &&
          assignment.assignmentSource === assignmentSource &&
          assignment.status === 'active'
      )
      .map((assignment) => ({
        ...assignment,
        permissionProfileIds: assignment.permissionProfileIds.filter((id) =>
          permissionProfileIds.includes(id)
        ),
      }))
      .filter((assignment) => assignment.permissionProfileIds.length > 0);
    const retainedIds = new Set(retained.flatMap((assignment) => assignment.permissionProfileIds));
    const addedIds = permissionProfileIds.filter((id) => !retainedIds.has(id));
    if (addedIds.length > 0) {
      retained.push({
        appId,
        permissionProfileIds: addedIds,
        assignmentSource,
        status: 'active',
        scopes: [],
      });
    }
    return retained;
  });
}
