import type { User } from '@voxire/auth';

/**
 * Treats a user as an admin/owner for the purposes of the license notice.
 *
 * The notice is operator-facing, so only owners, admins, and console admins
 * ever see it. Regular staff/viewers never do.
 */
export function isLicenseAdmin(user: User | null, role: string | null): boolean {
  const effectiveRole = user?.role ?? role ?? null;
  if (effectiveRole === 'owner' || effectiveRole === 'admin') return true;
  if (user?.isConsoleAdmin === true) return true;
  if (user?.is_console_admin === true) return true;
  return false;
}
