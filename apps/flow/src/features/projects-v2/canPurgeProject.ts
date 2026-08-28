import type { Project } from '@voxire/voxie/flow/project/v2/types';
import type { EffectiveAccess } from '@voxire/auth';
import { hasFlowPermission } from '../../shared/lib/roleUtils';

/**
 * Whether `userId` may permanently delete `project`, matching the server's gate
 * in `PurgeProjectV2`: central Flow management permission, the project's
 * creator, or its owner.
 *
 * Fails closed on missing access data or user id, since the action is
 * irreversible.
 */
export function canPurgeProject(
  project: Pick<Project, 'createdBy' | 'ownerId'>,
  userId: string | null,
  effectiveAccess: EffectiveAccess | null
): boolean {
  if (hasFlowPermission(effectiveAccess, 'flow:settings:manage')) return true;
  if (userId === null || userId === '') return false;
  // Both ids are `string | undefined` in the generated contract (int64 travels
  // as a wire string), so an absent field simply fails to match rather than
  // needing a coercion.
  return project.createdBy === userId || project.ownerId === userId;
}
