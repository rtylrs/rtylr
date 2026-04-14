import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { hrApi, type PayrollRow } from '../../shared/services/hrApi';

const payrollRowSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  role: z.string(),
  hoursWorked: z.number(),
  basePay: z.number(),
  overtime: z.number(),
  deductions: z.number(),
  status: z.enum(['Pending', 'Paid']).optional(),
});

/** Legacy hook used by Operations screen — reads from old payroll endpoint */
export function usePayroll(params?: { period?: string }) {
  return useQuery<PayrollRow[]>({
    queryKey: ['payroll-legacy', params],
    queryFn: async (): Promise<PayrollRow[]> => {
      const res = await unwrap(hrApi.listPayroll(params));
      const raw = res ?? [];
      return z.array(payrollRowSchema).parse(Array.isArray(raw) ? raw : []) as PayrollRow[];
    },
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
  PayrollRunEntry,
  ListPayrollRunsParams,
} from '../../shared/hooks/queries/usePayroll';
