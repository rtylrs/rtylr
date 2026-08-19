import { can } from '@voxire/permissions';
import type { PlatformRole, Resource } from '@voxire/permissions';

export type InsightsPermission = Extract<Resource, `insights:${string}`>;

// Typed against the full PlatformRole set (not @voxire/auth's narrower
// UserRole) because the permission matrix distinguishes billing,
// menu_manager, and store_manager from the app-access roles, and export
// gating must hold for all of them, not just the roles the auth store
// currently models. A UserRole from useUserStore is always a valid
// PlatformRole, so call sites are unaffected.
export function canInsightsPermission(
  role: PlatformRole | null | undefined,
  permission: InsightsPermission
): boolean {
  return role ? can({ platformRole: role }, permission) : false;
}
