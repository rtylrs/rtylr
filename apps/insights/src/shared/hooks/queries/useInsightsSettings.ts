import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';
import type { InsightsSettingsResponse, ScheduledReportsResponse, ScheduledReport } from './types';
import {
  mapInsightsSettingsResponse,
  mapScheduledReport,
  mapScheduledReportsResponse,
  toScheduledReportPayload,
} from '../../utils/onekitMappers';

export const useInsightsSettings = (): UseQueryResult<InsightsSettingsResponse> => {
  return useQuery<InsightsSettingsResponse>({
    queryKey: ['insights-settings'],
    queryFn: async (): Promise<InsightsSettingsResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/setting/v1/settings')
      );
      return mapInsightsSettingsResponse(res as Record<string, unknown>);
    },
  });
};

export const useUpdateInsightsSettings = (): UseMutationResult<
  InsightsSettingsResponse,
  unknown,
  Partial<InsightsSettingsResponse>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<InsightsSettingsResponse, unknown, Partial<InsightsSettingsResponse>>({
    mutationFn: async (
      data: Partial<InsightsSettingsResponse>
    ): Promise<InsightsSettingsResponse> => {
      const res = await unwrap(
        safe.put<Record<string, unknown>>('/api/insights/setting/v1/settings', {
          settings: {
            default_period: 'this_month',
            currency: data.defaultCurrency ?? 'USD',
            compare_previous: false,
            dashboard_widgets: [],
            favorite_reports: [],
          },
        })
      );
      return mapInsightsSettingsResponse(res as Record<string, unknown>);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['insights-settings'] });
      toast.success(t('insights.settings.saved'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};

export const useScheduledReports = (): UseQueryResult<ScheduledReportsResponse> => {
  return useQuery<ScheduledReportsResponse>({
    queryKey: ['insights-scheduled-reports'],
    queryFn: async (): Promise<ScheduledReportsResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/report/v1/scheduled')
      );
      return mapScheduledReportsResponse(res as Record<string, unknown>);
    },
  });
};

export interface CreateScheduledReportPayload {
  templateId: string;
  cadence: string;
  channel: string;
  recipient: string;
  isActive: boolean;
}

export const useCreateScheduledReport = (): UseMutationResult<
  ScheduledReport,
  unknown,
  CreateScheduledReportPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<ScheduledReport, unknown, CreateScheduledReportPayload>({
    mutationFn: async (data: CreateScheduledReportPayload): Promise<ScheduledReport> => {
      const res = await unwrap(
        safe.post<Record<string, unknown>>('/api/insights/report/v1/scheduled', toScheduledReportPayload(data))
      );
      const report = (res as Record<string, unknown>).report as Record<string, unknown>;
      return mapScheduledReport(report);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['insights-scheduled-reports'] });
      toast.success(t('insights.settings.scheduledReports.created'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};

export interface UpdateScheduledReportPayload {
  id: number;
  data: CreateScheduledReportPayload;
}

export const useUpdateScheduledReport = (): UseMutationResult<
  ScheduledReport,
  unknown,
  UpdateScheduledReportPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<ScheduledReport, unknown, UpdateScheduledReportPayload>({
    mutationFn: async ({ id, data }: UpdateScheduledReportPayload): Promise<ScheduledReport> => {
      const res = await unwrap(
        safe.put<Record<string, unknown>>(`/api/insights/report/v1/scheduled/${id}`, {
          ...toScheduledReportPayload(data),
        })
      );
      const report = (res as Record<string, unknown>).report as Record<string, unknown>;
      return mapScheduledReport(report);
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['insights-scheduled-reports'] });
      toast.success(t('insights.settings.scheduledReports.updated'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};

export const useDeleteScheduledReport = (): UseMutationResult<
  { deleted: boolean },
  unknown,
  number
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<{ deleted: boolean }, unknown, number>({
    mutationFn: async (id: number): Promise<{ deleted: boolean }> => {
      const res = await unwrap(
        safe.delete<{ deleted: boolean }>(`/api/insights/report/v1/scheduled/${id}`)
      );
      return res as { deleted: boolean };
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['insights-scheduled-reports'] });
      toast.success(t('insights.settings.scheduledReports.deleted'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};
