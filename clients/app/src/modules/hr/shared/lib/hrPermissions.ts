import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey } from '@voxire/types';
import { useUserStore } from '../../../../shared/stores/userStore';

export type HRPermission = PermissionKey &
  (`hr:${string}` | `hub:${string}` | `insights:${string}`);

function storedEffectiveAccess(): EffectiveAccess | undefined {
  const state = useUserStore.getState();
  return state.user?.effectiveAccess ?? undefined;
}

export function canHrPermission(
  permission: HRPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  return canEffectiveAccess(effectiveAccess, permission);
}

export function canViewCompensation(
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  return (
    canHrPermission('hr:employees:sensitive_view', effectiveAccess) ||
    canHrPermission('hr:payroll:view', effectiveAccess)
  );
}
