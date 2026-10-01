import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey } from '@voxire/types';
import { useUserStore } from '@voxire/auth';

export type InsightsPermission = PermissionKey & `insights:${string}`;

function storedEffectiveAccess(): EffectiveAccess | undefined {
  const state = useUserStore.getState();
  return state.user?.effectiveAccess ?? state.effectiveAccess ?? undefined;
}

export function canInsightsPermission(
  permission: InsightsPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  return canEffectiveAccess(effectiveAccess, permission);
}
