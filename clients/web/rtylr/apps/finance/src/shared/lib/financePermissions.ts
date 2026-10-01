import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey } from '@voxire/types';
import { useUserStore } from '@voxire/auth';

export type FinancePermission = PermissionKey & (`finance:${string}` | `hub:team:${string}`);

function storedEffectiveAccess(): EffectiveAccess | undefined {
  const state = useUserStore.getState();
  return state.user?.effectiveAccess ?? state.effectiveAccess ?? undefined;
}

export function canFinancePermission(
  permission: FinancePermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  return canEffectiveAccess(effectiveAccess, permission);
}
