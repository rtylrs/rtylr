import type { User } from '@voxire/auth';

/**
 * Treats a user as an admin/owner for the purposes of the license notice.
 *
 * The notice is operator-facing, so only owners, admins, and console admins
 * ever see it. Regular staff/viewers never do.
 */
export function isLicenseAdmin(user: User | null): boolean {
  return (
    user?.effectiveAccess?.isOwner === true ||
    user?.effectiveAccess?.isAdmin === true ||
    user?.isConsoleAdmin === true
  );
}
