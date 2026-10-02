import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey } from '@voxire/types';
import type { FlowProjectRole } from '../services/flowApi';

/** Flow-wide permissions use the server-resolved authorization document. */
export function hasFlowPermission(
  effectiveAccess: EffectiveAccess | null | undefined,
  permission: PermissionKey
): boolean {
  return canEffectiveAccess(effectiveAccess, permission, 'flow');
}

export function canFlowPermission(
  permission: PermissionKey,
  effectiveAccess: EffectiveAccess | null | undefined
): boolean {
  return hasFlowPermission(effectiveAccess, permission);
}

/** These helpers apply only to project-scoped collaborators. */
export function canEditTask(role: FlowProjectRole): boolean {
  return role !== 'viewer';
}

export function canDelete(role: FlowProjectRole): boolean {
  return role === 'admin';
}

const PROJECT_ROLE_RANK: Record<FlowProjectRole, number> = {
  admin: 0,
  member: 1,
  viewer: 2,
};

/** Returns the least privileged role across a set of project collaborators. */
export function mostRestrictiveProjectRole(roles: FlowProjectRole[]): FlowProjectRole | undefined {
  return roles.reduce<FlowProjectRole | undefined>(
    (lowest, role) =>
      lowest === undefined || PROJECT_ROLE_RANK[role] > PROJECT_ROLE_RANK[lowest] ? role : lowest,
    undefined
  );
}

export function projectRoleLabel(role: FlowProjectRole): string {
  const labels: Record<FlowProjectRole, string> = {
    admin: 'flowCommon.roles.admin',
    member: 'flowCommon.roles.member',
    viewer: 'flowCommon.roles.viewer',
  };
  return labels[role];
}
