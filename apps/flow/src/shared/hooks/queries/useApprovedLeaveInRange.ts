import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { LeaveServiceClient } from '@voxire/voxie/hr/leave/v1/client';

// HR is a foreign domain: read-only approved-leave lookup so Flow can warn
// before assigning work into someone's time off. Client app id stays 'flow'.
const leaveClient = createOneKitClient(LeaveServiceClient, 'flow');

export interface ApprovedLeaveOverlap {
  /** HR employee id (hr_employees.id). Mapping to hub users is pending
   *  rtylr#848, so callers surface ranges, not names. */
  employeeId?: string;
  startDate?: string;
  endDate?: string;
}

export const APPROVED_LEAVE_RANGE_KEY = (from: string, to: string) =>
  ['hr', 'leave', 'approved-range', { from, to }] as const;

/**
 * Approved leave overlapping [from, to] (server applies the contract's overlap
 * semantics: start <= to AND end >= from).
 *
 * Degrade (#838): a tenant without HR — or a failed read — yields null and the
 * warning hides; assignment itself is never blocked.
 */
export function useApprovedLeaveInRange(from: string, to: string, enabled = true) {
  return useQuery<ApprovedLeaveOverlap[] | null>({
    queryKey: APPROVED_LEAVE_RANGE_KEY(from, to),
    queryFn: async (): Promise<ApprovedLeaveOverlap[] | null> => {
      try {
        const res = await leaveClient.listLeaveRequests({
          status: 'approved',
          fromDate: from,
          toDate: to,
          limit: 100,
          offset: 0,
        });
        if (res === null || res === undefined) return null;
        return (res.requests ?? []).map((r) => ({
          employeeId: r.employeeId,
          startDate: r.startDate,
          endDate: r.endDate,
        }));
      } catch {
        return null;
      }
    },
    enabled,
    staleTime: 2 * 60 * 1000,
    retry: false,
  });
}
