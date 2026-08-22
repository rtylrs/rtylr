import type { Project } from '@voxire/voxie/flow/project/v2/types';
import type { UserRole } from '@voxire/auth';

/**
 * Business-wide roles that may purge any project in the bin.
 *
 * This mirrors onion's `hubProfileBypasses`
 * (internal/modules/workspaces/internal/application/service.go), which admits
 * `IsOwner || role == "admin" || role == "operator"`. Keep the two in step: the
 * server is the authority and rejects anyone it disagrees with, so a client that
 * drifts wider shows a button that 403s, and one that drifts narrower hides a
 * capability the user actually has.
 */
const PURGE_BYPASS_ROLES: readonly UserRole[] = ['owner', 'admin', 'operator'];

/**
 * Whether `userId` may permanently delete `project`, matching the server's gate
 * in `PurgeProjectV2`: a bypass role, the project's creator, or its owner.
 *
 * Fails closed on a missing user id or an unrecognised role rather than
 * defaulting to visible, since the action is irreversible.
 */
export function canPurgeProject(
  project: Pick<Project, 'createdBy' | 'ownerId'>,
  userId: string | null,
  role: UserRole | null
): boolean {
  if (role !== null && PURGE_BYPASS_ROLES.includes(role)) return true;
  if (userId === null || userId === '') return false;
  // Both ids are `string | undefined` in the generated contract (int64 travels
  // as a wire string), so an absent field simply fails to match rather than
  // needing a coercion.
  return project.createdBy === userId || project.ownerId === userId;
}
