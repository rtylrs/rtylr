import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { EmployeeServiceClient } from '@voxire/voxie/hr/employee/v1/client';
import { WorkScheduleServiceClient } from '@voxire/voxie/hr/work_schedule/v1/client';

const employeeClient = createOneKitClient(EmployeeServiceClient, 'flow');
const workScheduleClient = createOneKitClient(WorkScheduleServiceClient, 'flow');

/**
 * Contracted hours per week, keyed by the hub user id Flow identifies members
 * by.
 *
 * Flow used to set `capacity: 40` on every member because HR's
 * `contracted_hours_per_week` had no read path (#868). It does now, but reaching
 * it takes two hops: Flow knows hub user ids, `hr.employee.v1` maps `user_id` to
 * `work_schedule_id`, and `hr.work_schedule.v1` maps that to the hours.
 *
 * A member absent from this map has genuinely unknown capacity, and callers must
 * render it as unknown. Deliberately not defaulted: a part-timer shown against
 * 40 hours reads as under-loaded, and a four-day week reads as a shortfall,
 * which is the bug this replaces. An unknown denominator must not silently
 * become a plausible one.
 */
export function useMemberCapacityHours() {
  return useQuery({
    queryKey: ['member-capacity-hours'],
    queryFn: async (): Promise<Record<string, number>> => {
      const [employeesResponse, schedulesResponse] = await Promise.all([
        // Bounded, per CLAUDE.md. Both are per-business reference lists rather
        // than growing feeds, so one page is the whole set in practice.
        employeeClient.listEmployees({ limit: 500, offset: 0 }),
        workScheduleClient.listWorkSchedules({ limit: 500, offset: 0 }),
      ]);

      const hoursByScheduleId = new Map<string, number>();
      for (const schedule of schedulesResponse.schedules ?? []) {
        // Zero or absent means the schedule does not record contracted hours.
        // Skipping keeps it unknown rather than turning it into "0 hours", which
        // would render as permanently over capacity.
        if (!schedule.contractedHoursPerWeek || schedule.contractedHoursPerWeek <= 0) continue;
        hoursByScheduleId.set(String(schedule.id), schedule.contractedHoursPerWeek);
      }

      const hoursByUserId: Record<string, number> = {};
      for (const employee of employeesResponse.employees ?? []) {
        if (!employee.userId || !employee.workScheduleId) continue;
        const hours = hoursByScheduleId.get(String(employee.workScheduleId));
        if (hours === undefined) continue;
        hoursByUserId[String(employee.userId)] = hours;
      }

      return hoursByUserId;
    },
  });
}
