import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { canEffectiveAccess } from '@voxire/permissions';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { EmployeeServiceClient } from '@voxire/voxie/hr/employee/v1/client';
import { WorkScheduleServiceClient } from '@voxire/voxie/hr/work_schedule/v1/client';

const employeeClient = createOneKitClient(EmployeeServiceClient, 'hr');
const workScheduleClient = createOneKitClient(WorkScheduleServiceClient, 'hr');

/** Missing contracted hours stay unknown rather than defaulting to 40. */
export function useMemberCapacityHours() {
  const access = useUserStore((state) => state.user?.effectiveAccess);
  return useQuery({
    enabled:
      canEffectiveAccess(access, 'hr:employees:view', 'hr') &&
      canEffectiveAccess(access, 'hr:schedules:manage', 'hr'),
    queryKey: ['member-capacity-hours'],
    queryFn: async (): Promise<Record<string, number>> => {
      const [employeesResponse, schedulesResponse] = await Promise.all([
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
