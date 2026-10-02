import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  PayrollRunPayload,
  PayrollEntryPayload,
} from '@voxire/voxie/hr/employees/payroll/runs/v1/types';
import type { TriggerPayrollRunRequest } from '@voxire/voxie/hr/employees/payroll/runs/v1/types';
import type { ListPayrollRunsResponse } from '@voxire/voxie/hr/employees/payroll/runs/v1/types';
import { hrTransport } from '../../lib/clients';
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
  period: 'weekly' | 'biweekly' | 'monthly';
  overtimeEnabled: boolean;
  payDayOfMonth: number;
}

export interface UpsertPayrollConfigRequest {
  period: 'weekly' | 'biweekly' | 'monthly';
  overtimeEnabled: boolean;
  payDayOfMonth: number;
}

interface ListPayrollRunsParams {
  status?: string;
  limit?: number;
  offset?: number;
}

interface PayrollQueryOptions {
  enabled?: boolean;
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

/* ── Config hooks ── */

export const useGetPayrollConfig = () => {
  return useQuery<PayrollConfigData>({
    queryKey: ['hr', 'payroll-config'],
    queryFn: async (): Promise<PayrollConfigData> => {
      try {
        const res = await unwrap(hrTransport.get<Record<string, unknown>>('/hr/payroll/config'));
        const raw = (res?.config || res || {}) as Record<string, unknown>;
        return {
          period:
            ((raw.period ?? raw.periodType ?? raw.period_type) as PayrollConfigData['period']) ??
            'monthly',
          overtimeEnabled: Boolean(raw.overtimeEnabled ?? raw.overtime_enabled ?? false),
          payDayOfMonth: Number(raw.payDayOfMonth ?? raw.pay_day_of_month ?? 28),
        };
      } catch (error) {
        const status =
          (error as { response?: { status?: number }; status?: number })?.response?.status ??
          (error as { status?: number })?.status;
        if (status !== 404) throw error;
        return {
          period: 'monthly',
          overtimeEnabled: false,
          payDayOfMonth: 28,
        };
      }
    },
  });
};

export const useUpsertPayrollConfig = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (data: UpsertPayrollConfigRequest): Promise<void> => {
      await unwrap(
        hrTransport.put('/hr/payroll/config', {
          periodType: data.period,
          payDayOfMonth: data.payDayOfMonth,
          overtimeEnabled: data.overtimeEnabled,
        })
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'payroll-config'] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'hr-config'] });
      toast.success(t('hrPayroll.configSaved'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};

/* ── Run hooks ── */

export const useListPayrollRuns = (
  params?: ListPayrollRunsParams,
  options?: PayrollQueryOptions
) => {
  return useQuery<ListPayrollRunsResponse>({
    queryKey: ['hr', 'payroll-runs', params],
    queryFn: async (): Promise<ListPayrollRunsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListPayrollRunsResponse>('/hr/payroll/runs', { params })
      );
      return res as ListPayrollRunsResponse;
    },
    enabled: options?.enabled ?? true,
  });
};

export const useGetPayrollRun = (id: number | undefined, options?: PayrollQueryOptions) => {
  return useQuery<PayrollRunPayload>({
    queryKey: ['hr', 'payroll-run', id],
    queryFn: async (): Promise<PayrollRunPayload> => {
      const res = await unwrap(
        hrTransport.get<{ run: PayrollRunPayload }>(`/hr/payroll/runs/${id}`)
      );
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
    queryKey: ['hr', 'my-payroll-runs', params],
    queryFn: async (): Promise<ListPayrollRunsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListPayrollRunsResponse>('/hr/payroll/my-runs', { params })
      );
      return res as ListPayrollRunsResponse;
    },
    enabled: options?.enabled ?? true,
    refetchInterval: 60_000,
    refetchOnWindowFocus: 'always',
  });
};

/**
 * Returns the run with `entries` containing exactly the caller's own entry.
 * A run that isn't theirs, or isn't approved/exported, comes back as a 404.
 */
export const useGetMyPayrollRun = (id: number | undefined, options?: PayrollQueryOptions) => {
  return useQuery<PayrollRunPayload>({
    queryKey: ['hr', 'my-payroll-run', id],
    queryFn: async (): Promise<PayrollRunPayload> => {
      const res = await unwrap(
        hrTransport.get<{ run: PayrollRunPayload }>(`/hr/payroll/my-runs/${id}`)
      );
      return (res as { run: PayrollRunPayload }).run;
    },
    enabled: !!id && (options?.enabled ?? true),
    refetchInterval: 60_000,
    refetchOnWindowFocus: 'always',
  });
};

export const useTriggerPayrollRun = (enabled = true) => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (data: TriggerPayrollRunRequest): Promise<void> => {
      if (!enabled) return;
      await unwrap(hrTransport.post('/hr/payroll/runs/trigger', data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'payroll-runs'] });
      toast.success(t('hrPayroll.runTriggered'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};

/**
 * Posts an approved payroll run to the central finance journal
 * (HR->FIN-01). onion's POST /hr/payroll/runs/{id}/export takes an
 * idempotent claim, converts internally-stored cents at the boundary and
 * posts through the Finance-owned financejournal port, using this tenant's
 * configured payroll account mapping. Re-posting an exported run is a
 * server-side no-op error, so the button only shows for approved runs.
 */
export const useExportPayrollRun = (enabled = true) => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (runId: number): Promise<void> => {
      if (!enabled) return;
      await unwrap(hrTransport.post(`/hr/payroll/runs/${runId}/export`));
    },
    onSuccess: (_data: void, runId: number) => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'payroll-runs'] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'payroll-run', runId] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'my-payroll-runs'] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'my-payroll-run', runId] });
      toast.success(t('hrPayroll.runDetail.posted'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};

export const useApprovePayrollRun = (enabled = true) => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (data: ApprovePayrollRunRequest): Promise<void> => {
      if (!enabled) return;
      await unwrap(hrTransport.post(`/hr/payroll/runs/${data.runId}/approve`, data));
    },
    onSuccess: (_data: void, variables: ApprovePayrollRunRequest) => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'payroll-runs'] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'payroll-run', variables.runId] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'my-payroll-runs'] });
      queryClient.invalidateQueries({ queryKey: ['hr', 'my-payroll-run', variables.runId] });
      toast.success(t('hrPayroll.runApproved'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};
