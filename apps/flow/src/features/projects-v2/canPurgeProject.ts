import type { Project } from '@voxire/voxie/flow/project/v2/types';
import type { UserRole } from '@voxire/auth';
import { hasBusinessRoleBypass } from '../../shared/lib/roleUtils';

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
  if (hasBusinessRoleBypass(role)) return true;
  if (userId === null || userId === '') return false;
  // Both ids are `string | undefined` in the generated contract (int64 travels
  // as a wire string), so an absent field simply fails to match rather than
  // needing a coercion.
  return project.createdBy === userId || project.ownerId === userId;
}
