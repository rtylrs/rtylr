import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  PayrollRunPayload,
  PayrollEntryPayload,
} from '@voxire/contracts/hr/employees/payroll/runs/v1/types';
import type { TriggerPayrollRunRequest } from '@voxire/contracts/hr/employees/payroll/runs/v1/trigger_payroll_run';
import type { ApprovePayrollRunRequest } from '@voxire/contracts/hr/employees/payroll/runs/v1/approve_payroll_run';
import type { ExportPayrollRequest } from '@voxire/contracts/hr/employees/payroll/runs/v1/export_payroll';
import type { ListPayrollRunsResponse } from '@voxire/contracts/hr/employees/payroll/runs/v1/list_payroll_runs';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

/* ── Local config type (no contract exists) ── */

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

interface ListPayrollRunsParams {
  status?: string;
  limit?: number;
  offset?: number;
}

export type { PayrollConfigData, PayrollRunPayload, PayrollEntryPayload, ListPayrollRunsParams };

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
      const res = await unwrap(
        safe.put<{ config: PayrollConfigData }>('/hr/payroll/config', {
          periodType: data.period,
          overtimeEnabled: data.overtimeEnabled,
        })
      );
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
      const res = await unwrap(safe.get<{ run: PayrollRunPayload }>(`/hr/payroll/runs/${id}`));
      return (res as { run: PayrollRunPayload }).run;
    },
    enabled: !!id,
  });
};

export const useTriggerPayrollRun = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: TriggerPayrollRunRequest): Promise<void> => {
      await unwrap(safe.post('/hr/payroll/runs/trigger', data));
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
      await unwrap(safe.post(`/hr/payroll/runs/${data.runId}/approve`, data));
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
