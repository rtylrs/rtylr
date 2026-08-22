import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { PayrollServiceClient } from '@voxire/voxie/hr/payroll/v1/client';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
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
  enabled?: boolean;
}): ReturnType<typeof useQuery<PayrollRow[]>> {
  const enabled = params?.enabled ?? true;
  return useQuery<PayrollRow[]>({
    queryKey: ['payroll-legacy', params?.period ?? '', enabled],
    queryFn: async (): Promise<PayrollRow[]> => {
      const [payrollRes, staffRes] = await Promise.all([
        payrollClient.listPayroll({ periodLabel: params?.period ?? '' }),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);
      const staffById = new Map((staffRes.staff ?? []).map((s) => [Number(s.id), s]));

      return (payrollRes.records ?? []).map((r) => {
        const staff = staffById.get(r.staffId ?? 0);
        return {
          id: r.id ?? 0,
          name: staff ? `${staff.firstName} ${staff.lastName}`.trim() : `Staff #${r.staffId}`,
          role: staff?.role ?? '',
          hoursWorked: (r.regularHours ?? 0) + (r.overtimeHours ?? 0),
          basePay: r.basePay ?? 0,
          overtime: r.overtimePay ?? 0,
          deductions: r.deductions ?? 0,
          status: normalizeStatus(r.status ?? ''),
        };
      });
    },
    enabled,
  });
}

export {
  useListPayrollRuns,
  useGetPayrollRun,
  useTriggerPayrollRun,
  useApprovePayrollRun,
  useExportPayrollRun,
  useGetPayrollConfig,
  useUpsertPayrollConfig,
} from '../../shared/hooks/queries/usePayroll';

export type {
  PayrollConfigData,
  PayrollRunPayload,
  PayrollEntryPayload,
  ListPayrollRunsParams,
} from '../../shared/hooks/queries/usePayroll';
