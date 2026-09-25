import { can, canEffectiveAccess, canonicalizePermission } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';
import { useUserStore } from '@voxire/auth';
import type { EffectiveAccess } from '@voxire/types';

export type FinancePermission = Extract<
  Resource,
  `finance:${string}` | 'hub:team:invite' | 'hub:team:remove'
>;

function storedEffectiveAccess(): EffectiveAccess | undefined {
  return typeof useUserStore.getState === 'function'
    ? useUserStore.getState().user?.effectiveAccess
    : undefined;
}

export function canFinancePermission(
  role: UserRole | null | undefined,
  permission: FinancePermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  if (effectiveAccess)
    return canEffectiveAccess(effectiveAccess, canonicalizePermission(permission));
  return role ? can({ platformRole: role }, permission) : false;
}
