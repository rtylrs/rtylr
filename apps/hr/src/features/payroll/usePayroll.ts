import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { PayrollServiceClient } from '@voxire/voxie/hr/payroll/v1/service_client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/service_client';
import { type PayrollRow } from '../../shared/services/hrApi';

const payrollClient = createOneKitClient(PayrollServiceClient, 'hr');
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

function normalizeStatus(status: string): 'Pending' | 'Paid' | undefined {
  const lower = status.toLowerCase();
  if (lower === 'paid') return 'Paid';
  if (lower === 'pending') return 'Pending';
  return undefined;
}

/** Legacy hook used by Operations screen, reads from hr.payroll.v1.ListPayroll */
export function usePayroll(params?: {
  period?: string;
}): ReturnType<typeof useQuery<PayrollRow[]>> {
  return useQuery<PayrollRow[]>({
    queryKey: ['payroll-legacy', params],
    queryFn: async (): Promise<PayrollRow[]> => {
      const [payrollRes, staffRes] = await Promise.all([
        payrollClient.listPayroll({ periodLabel: params?.period ?? '' }),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const staffById = new Map(staffRes.staff.map((s) => [Number(s.id), s]));

      return payrollRes.records.map((r) => {
        const staff = staffById.get(r.staffId);
        return {
          id: r.id,
          name: staff ? `${staff.firstName} ${staff.lastName}`.trim() : `Staff #${r.staffId}`,
          role: staff?.role ?? '',
          hoursWorked: r.regularHours + r.overtimeHours,
          basePay: r.basePay,
          overtime: r.overtimePay,
          deductions: r.deductions,
          status: normalizeStatus(r.status),
        };
      });
    },
  });
}

export {
  useListPayrollRuns,
  useGetPayrollRun,
  useTriggerPayrollRun,
  useApprovePayrollRun,
  useGetPayrollConfig,
  useUpsertPayrollConfig,
} from '../../shared/hooks/queries/usePayroll';

export type {
  PayrollConfigData,
  PayrollRunPayload,
  PayrollEntryPayload,
  ListPayrollRunsParams,
} from '../../shared/hooks/queries/usePayroll';
