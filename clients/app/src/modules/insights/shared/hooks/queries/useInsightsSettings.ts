import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { SettingServiceClient } from '@voxire/voxie/insights/setting/v1/client';
import type { InsightsSettings as ApiInsightsSettings } from '@voxire/voxie/insights/setting/v1/types';
import { ReportServiceClient } from '@voxire/voxie/insights/report/v1/client';
import type { ScheduledReport as ApiScheduledReport } from '@voxire/voxie/insights/report/v1/types';
import type { InsightsSettingsResponse, ScheduledReportsResponse, ScheduledReport } from './types';

const settingClient = createOneKitClient(SettingServiceClient, 'insights');
const reportClient = createOneKitClient(ReportServiceClient, 'insights');

const DEFAULT_API_SETTINGS: ApiInsightsSettings = {
  defaultPeriod: '',
  currency: 'USD',
  fiscalYearStart: '01-01',
  timezone: 'UTC',
  comparePrevious: false,
  dashboardWidgets: [],
  favoriteReports: [],
};

function fromApiSettings(settings: ApiInsightsSettings): InsightsSettingsResponse {
  return {
    defaultCurrency: settings.currency ?? 'USD',
    defaultPeriod: settings.defaultPeriod ?? '',
    comparePrevious: settings.comparePrevious ?? false,
    dashboardWidgets: settings.dashboardWidgets ?? [],
    favoriteReports: settings.favoriteReports ?? [],
    fiscalYearStart: settings.fiscalYearStart || '01-01',
    timezone: settings.timezone || 'UTC',
  };
}

function fromApiScheduledReport(report: ApiScheduledReport): ScheduledReport {
  return {
    id: report.id ?? '',
    templateId: report.templateName || report.templateId || '',
    cadence: report.frequency ?? '',
    channel: report.channel ?? '',
    recipient: report.recipientEmail ?? '',
    isActive: report.isActive ?? false,
    createdAt: report.createdAt ?? '',
  };
}

export const useInsightsSettings = (): UseQueryResult<InsightsSettingsResponse> => {
  return useQuery<InsightsSettingsResponse>({
    queryKey: ['insights-settings'],
    queryFn: async (): Promise<InsightsSettingsResponse> => {
      const res = await settingClient.getInsightsSettings({});
      return fromApiSettings(res.settings ?? DEFAULT_API_SETTINGS);
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
      // insights.setting.v1.UpdateInsightsSettings replaces the whole
      // settings object, so merge onto the currently cached settings rather
      // than sending only the changed fields.
      const current = qc.getQueryData<InsightsSettingsResponse>(['insights-settings']);
      const res = await settingClient.updateInsightsSettings({
        settings: {
          defaultPeriod:
            data.defaultPeriod ?? current?.defaultPeriod ?? DEFAULT_API_SETTINGS.defaultPeriod,
          currency:
            data.defaultCurrency ?? current?.defaultCurrency ?? DEFAULT_API_SETTINGS.currency,
          fiscalYearStart:
            data.fiscalYearStart ??
            current?.fiscalYearStart ??
            DEFAULT_API_SETTINGS.fiscalYearStart,
          timezone: data.timezone ?? current?.timezone ?? DEFAULT_API_SETTINGS.timezone,
          comparePrevious:
            data.comparePrevious ??
            current?.comparePrevious ??
            DEFAULT_API_SETTINGS.comparePrevious,
          dashboardWidgets:
            data.dashboardWidgets ??
            current?.dashboardWidgets ??
            DEFAULT_API_SETTINGS.dashboardWidgets,
          favoriteReports:
            data.favoriteReports ??
            current?.favoriteReports ??
            DEFAULT_API_SETTINGS.favoriteReports,
        },
      });
      return fromApiSettings(res.settings ?? DEFAULT_API_SETTINGS);
    },
    onSuccess: (settings): void => {
      qc.setQueryData(['insights-settings'], settings);
      void qc.invalidateQueries({ queryKey: ['insights-settings'] });
      toast.success(t('insights.settings.saved'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};

export const useScheduledReports = ({
  enabled = true,
} = {}): UseQueryResult<ScheduledReportsResponse> => {
  return useQuery<ScheduledReportsResponse>({
    queryKey: ['insights-scheduled-reports'],
    enabled,
    queryFn: async (): Promise<ScheduledReportsResponse> => {
      const res = await reportClient.listScheduledReports({
        page: 1,
        limit: 100,
        sortBy: 'createdAt',
        sortDir: 'desc',
      });
      return { scheduledReports: (res.reports ?? []).map(fromApiScheduledReport) };
    },
  });
};

export const useDeleteScheduledReport = (): UseMutationResult<
  { deleted: boolean },
  unknown,
  string
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<{ deleted: boolean }, unknown, string>({
    mutationFn: async (id: string): Promise<{ deleted: boolean }> => {
      const res = await reportClient.deleteScheduledReport({ id });
      return { deleted: res.deleted ?? false };
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['insights-scheduled-reports'] });
      toast.success(t('insights.settings.scheduledReports.deleted'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('insights.errors.generic')));
    },
  });
};
