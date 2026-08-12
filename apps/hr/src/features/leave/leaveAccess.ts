import type { ListLeaveRequestsRequest } from '@voxire/voxie/hr/leave/v1/types';

/**
 * Keep the browser request aligned with the caller's leave scope. The Onion
 * service remains the authoritative check, but ordinary users should not
 * request the company feed in the first place.
 */
export function getLeaveRequestScope(
  userId: string | null | undefined,
  canApprove: boolean
): ListLeaveRequestsRequest {
  return {
    employeeId: canApprove ? undefined : userId || undefined,
    status: undefined,
  };
}
