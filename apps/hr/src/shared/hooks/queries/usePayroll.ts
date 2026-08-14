import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  PayrollRunPayload,
  PayrollEntryPayload,
} from '@voxire/voxie/hr/employees/payroll/runs/v1/types';
import type { TriggerPayrollRunRequest } from '@voxire/voxie/hr/employees/payroll/runs/v1/types';
import type { ListPayrollRunsResponse } from '@voxire/voxie/hr/employees/payroll/runs/v1/types';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

/* ── Local config type (no contract exists) ── */

/**
 * The `hr.employees.payroll.runs.v1.RunsService` proto exists but is never
 * registered as an onekit service in onion (see hr_payroll/module.go - only
 * `RegisterRoutes` is implemented, no `RegisterAPI`). Approval is served by
 * the legacy chi route only, with a body shape that doesn't match the
 * generated `ApprovePayrollRunRequest` (which assumes a live onekit rpc).
 */
interface ApprovePayrollRunRequest {
  runId: number;
}

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

interface PayrollQueryOptions {
  enabled?: boolean;
}

export interface PayrollAdjustment {
  id: number;
  employeeId: number;
  kind: 'bonus' | 'deduction';
  label: string;
  amount: number;
  effectiveFrom: string;
  effectiveTo: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * The self-service endpoints take no `status`: onion restricts them to
 * approved/exported server-side and resolves the employee from the session,
 * so there is deliberately no way to widen the scope from the client.
 */
interface ListMyPayrollRunsParams {
  limit?: number;
  offset?: number;
}

export type {
  PayrollConfigData,
  PayrollRunPayload,
  PayrollEntryPayload,
  ListPayrollRunsParams,
  ListMyPayrollRunsParams,
};

interface ListPayrollAdjustmentsResponse {
  adjustments: PayrollAdjustment[];
}

export const useListPayrollAdjustments = (enabled = true) => {
  return useQuery<ListPayrollAdjustmentsResponse>({
    queryKey: ['payroll-adjustments'],
    queryFn: async () => unwrap(safe.get<ListPayrollAdjustmentsResponse>('/hr/payroll/adjustments')),
    enabled,
  });
};

export interface PayrollAdjustmentInput {
  employeeId: number;
  kind: 'bonus' | 'deduction';
  label: string;
  amount: number;
  effectiveFrom: string;
  effectiveTo?: string;
}

export const useCreatePayrollAdjustment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: PayrollAdjustmentInput) =>
      unwrap(safe.post<{ adjustment: PayrollAdjustment }>('/hr/payroll/adjustments', input)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['payroll-adjustments'] });
      toast.success('Payroll adjustment saved');
    },
    onError: () => toast.error('Failed to save payroll adjustment'),
  });
};

export const useDeletePayrollAdjustment = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(safe.delete<{ success: boolean }>(`/hr/payroll/adjustments/${id}`)),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['payroll-adjustments'] }),
    onError: () => toast.error('Failed to remove payroll adjustment'),
  });
};

/* ── Config hooks ── */

export const useGetPayrollConfig = () => {
  return useQuery<PayrollConfigData>({
    queryKey: ['payroll-config'],
    queryFn: async (): Promise<PayrollConfigData> => {
      try {
        const res = await unwrap(safe.get<Record<string, unknown>>('/hr/payroll/config'));
        const raw = ((res as any)?.config || res || {}) as Record<string, unknown>;
        return {
          grossModel:
            ((raw.grossModel ?? raw.gross_model) as PayrollConfigData['grossModel']) ?? 'both',
          period:
            ((raw.period ?? raw.periodType ?? raw.period_type) as PayrollConfigData['period']) ??
            'monthly',
          overtimeEnabled: Boolean(raw.overtimeEnabled ?? raw.overtime_enabled ?? false),
          deductionsEnabled: Boolean(raw.deductionsEnabled ?? raw.deductions_enabled ?? false),
          bonusesEnabled: Boolean(raw.bonusesEnabled ?? raw.bonuses_enabled ?? false),
        };
      } catch (error) {
        const status =
          (error as { response?: { status?: number }; status?: number })?.response?.status ??
          (error as { status?: number })?.status;
        if (status !== 404) throw error;
        return {
          grossModel: 'both',
          period: 'monthly',
          overtimeEnabled: false,
          deductionsEnabled: false,
          bonusesEnabled: false,
        };
      }
    },
  });
};

export const useUpsertPayrollConfig = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpsertPayrollConfigRequest): Promise<PayrollConfigData> => {
      const res = await unwrap(
        safe.put<{ config?: PayrollConfigData }>('/hr/payroll/config', {
          periodType: data.period,
          overtimeEnabled: data.overtimeEnabled,
          grossModel: data.grossModel,
          deductionsEnabled: data.deductionsEnabled,
          bonusesEnabled: data.bonusesEnabled,
        })
      );
      const raw = ((res as { config?: PayrollConfigData }).config ?? res) as Record<string, unknown>;
      return {
        grossModel: ((raw.grossModel ?? raw.gross_model) as PayrollConfigData['grossModel']) ?? 'both',
        period: ((raw.period ?? raw.periodType ?? raw.period_type) as PayrollConfigData['period']) ?? 'monthly',
        overtimeEnabled: Boolean(raw.overtimeEnabled ?? raw.overtime_enabled ?? false),
        deductionsEnabled: Boolean(raw.deductionsEnabled ?? raw.deductions_enabled ?? false),
        bonusesEnabled: Boolean(raw.bonusesEnabled ?? raw.bonuses_enabled ?? false),
      };
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

export const useListPayrollRuns = (
  params?: ListPayrollRunsParams,
  options?: PayrollQueryOptions
) => {
  return useQuery<ListPayrollRunsResponse>({
    queryKey: ['payroll-runs', params],
    queryFn: async (): Promise<ListPayrollRunsResponse> => {
      const res = await unwrap(safe.get<ListPayrollRunsResponse>('/hr/payroll/runs', { params }));
      return res as ListPayrollRunsResponse;
    },
    enabled: options?.enabled ?? true,
  });
};

export const useGetPayrollRun = (id: number | undefined, options?: PayrollQueryOptions) => {
  return useQuery<PayrollRunPayload>({
    queryKey: ['payroll-run', id],
    queryFn: async (): Promise<PayrollRunPayload> => {
      const res = await unwrap(safe.get<{ run: PayrollRunPayload }>(`/hr/payroll/runs/${id}`));
      return (res as { run: PayrollRunPayload }).run;
    },
    enabled: !!id && (options?.enabled ?? true),
  });
};

/* ── Self-service run hooks ── */

/**
 * Employee-scoped counterparts to `useListPayrollRuns` / `useGetPayrollRun`.
 * The admin hooks hit `/hr/payroll/runs`, which onion gates on the
 * `manage_payroll` permission and which returns every employee's runs: wrong
 * on both counts for "My Pay Slips". These read `/hr/payroll/my-runs`, which
 * resolves the caller's own employee id from the session.
 *
 * Their query keys are deliberately distinct from the admin hooks' so that
 * invalidating one never serves company-wide data into a self-service screen.
 */
export const useListMyPayrollRuns = (
  params?: ListMyPayrollRunsParams,
  options?: PayrollQueryOptions
) => {
  return useQuery<ListPayrollRunsResponse>({
    queryKey: ['my-payroll-runs', params],
    queryFn: async (): Promise<ListPayrollRunsResponse> => {
      const res = await unwrap(
        safe.get<ListPayrollRunsResponse>('/hr/payroll/my-runs', { params })
      );
      return res as ListPayrollRunsResponse;
    },
    enabled: options?.enabled ?? true,
  });
};

/**
 * Returns the run with `entries` containing exactly the caller's own entry.
 * A run that isn't theirs, or isn't approved/exported, comes back as a 404.
 */
export const useGetMyPayrollRun = (id: number | undefined, options?: PayrollQueryOptions) => {
  return useQuery<PayrollRunPayload>({
    queryKey: ['my-payroll-run', id],
    queryFn: async (): Promise<PayrollRunPayload> => {
      const res = await unwrap(safe.get<{ run: PayrollRunPayload }>(`/hr/payroll/my-runs/${id}`));
      return (res as { run: PayrollRunPayload }).run;
    },
    enabled: !!id && (options?.enabled ?? true),
  });
};

export const useTriggerPayrollRun = (enabled = true) => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: TriggerPayrollRunRequest): Promise<void> => {
      if (!enabled) return;
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

export const useApprovePayrollRun = (enabled = true) => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApprovePayrollRunRequest): Promise<void> => {
      if (!enabled) return;
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
