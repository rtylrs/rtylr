import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

/* ── Types ── */

interface PayrollConfigData {
  grossModel: 'hourly' | 'salary' | 'both';
  period: 'weekly' | 'biweekly' | 'monthly';
  overtimeEnabled: boolean;
  deductionsEnabled: boolean;
  bonusesEnabled: boolean;
}

interface UpsertPayrollConfigRequest {
  grossModel?: 'hourly' | 'salary' | 'both';
  period?: 'weekly' | 'biweekly' | 'monthly';
  overtimeEnabled?: boolean;
  deductionsEnabled?: boolean;
  bonusesEnabled?: boolean;
}

interface PayrollRunEntry {
  id: number;
  employeeId: number;
  employeeName: string;
  hoursWorked: number;
  basePay: number;
  overtimePay: number;
  deductions: number;
  bonuses: number;
  netPay: number;
}

interface PayrollRunPayload {
  id: number;
  periodStart: string;
  periodEnd: string;
  status: 'Draft' | 'Pending' | 'Approved' | 'Paid';
  totalNetPay: number;
  totalBasePay: number;
  totalOvertimePay: number;
  totalDeductions: number;
  employeeCount: number;
  createdAt: string;
  entries?: PayrollRunEntry[];
}

interface ListPayrollRunsParams {
  status?: string;
  limit?: number;
  offset?: number;
}

interface ListPayrollRunsResponse {
  runs: PayrollRunPayload[];
  total: number;
}

interface GetPayrollRunResponse {
  run: PayrollRunPayload;
}

interface TriggerPayrollRunRequest {
  periodStart: string;
  periodEnd: string;
}

interface ApprovePayrollRunRequest {
  runId: number;
  approved: boolean;
}

interface ExportPayrollRunRequest {
  runId: number;
  format: 'csv' | 'pdf';
}

export type { PayrollConfigData, PayrollRunPayload, PayrollRunEntry, ListPayrollRunsParams };

/* ── Config hooks ── */

export const useGetPayrollConfig = () => {
  return useQuery<PayrollConfigData>({
    queryKey: ['payroll-config'],
    queryFn: async (): Promise<PayrollConfigData> => {
      const res = await unwrap(safe.get<{ config: PayrollConfigData }>('/hr/payroll/config'));
      return (res as { config: PayrollConfigData }).config;
    },
  });
};

export const useUpsertPayrollConfig = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpsertPayrollConfigRequest): Promise<PayrollConfigData> => {
      const res = await unwrap(safe.put<{ config: PayrollConfigData }>('/hr/payroll/config', data));
      return (res as { config: PayrollConfigData }).config;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payroll-config'] });
      queryClient.invalidateQueries({ queryKey: ['hr-config'] });
      toast.success(t('payroll.configSaved'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

/* ── Run hooks ── */

export const useListPayrollRuns = (params?: ListPayrollRunsParams) => {
  return useQuery<ListPayrollRunsResponse>({
    queryKey: ['payroll-runs', params],
    queryFn: async (): Promise<ListPayrollRunsResponse> => {
      const res = await unwrap(safe.get<ListPayrollRunsResponse>('/hr/payroll/runs', { params }));
      return res as ListPayrollRunsResponse;
    },
  });
};

export const useGetPayrollRun = (id: number | undefined) => {
  return useQuery<PayrollRunPayload>({
    queryKey: ['payroll-run', id],
    queryFn: async (): Promise<PayrollRunPayload> => {
      const res = await unwrap(safe.get<GetPayrollRunResponse>(`/hr/payroll/runs/${id}`));
      return (res as GetPayrollRunResponse).run;
    },
    enabled: !!id,
  });
};

export const useTriggerPayrollRun = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: TriggerPayrollRunRequest): Promise<void> => {
      await unwrap(safe.post('/hr/payroll/runs', data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payroll-runs'] });
      toast.success(t('payroll.runTriggered'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApprovePayrollRun = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApprovePayrollRunRequest): Promise<void> => {
      await unwrap(safe.put(`/hr/payroll/runs/${data.runId}/approve`, data));
    },
    onSuccess: (_data: void, variables: ApprovePayrollRunRequest) => {
      queryClient.invalidateQueries({ queryKey: ['payroll-runs'] });
      queryClient.invalidateQueries({ queryKey: ['payroll-run', variables.runId] });
      toast.success(t('payroll.runApproved'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useExportPayrollRun = () => {
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ExportPayrollRunRequest): Promise<Blob> => {
      const res = await unwrap(
        safe.post<Blob>(`/hr/payroll/runs/${data.runId}/export`, { format: data.format })
      );
      return res as Blob;
    },
    onSuccess: () => {
      toast.success(t('payroll.exported'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
