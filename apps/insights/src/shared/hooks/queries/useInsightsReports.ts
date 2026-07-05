import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type { ReportTemplatesResponse, GenerateReportResult } from './types';

export const useReportTemplates = (): UseQueryResult<ReportTemplatesResponse> => {
  return useQuery<ReportTemplatesResponse>({
    queryKey: ['insights-report-templates'],
    queryFn: async (): Promise<ReportTemplatesResponse> => {
      const res = await unwrap(
        safe.get<ReportTemplatesResponse>('/api/insights/report/v1/templates')
      );
      return res as ReportTemplatesResponse;
    },
  });
};

interface GenerateRevenueReportVariables {
  format?: string;
  from?: string;
  to?: string;
}

export const useGenerateRevenueReport = (): UseMutationResult<
  GenerateReportResult,
  unknown,
  GenerateRevenueReportVariables
> => {
  const { t } = useTranslation();
  return useMutation<GenerateReportResult, unknown, GenerateRevenueReportVariables>({
    mutationFn: async (
      variables: GenerateRevenueReportVariables
    ): Promise<GenerateReportResult> => {
      const res = await unwrap(
        safe.post<GenerateReportResult>('/api/insights/report/v1/reports/revenue/generate', variables)
      );
      return res as GenerateReportResult;
    },
    onSuccess: (): void => {
      toast.success(t('insights.reports.generateQueued'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};

// useRevenueCSVExportURL builds the fully-qualified CSV export URL from the
// configured API base. The CSV endpoint streams text/csv and must bypass the
// JSON-aware safeFetch wrapper, so the UI anchors to this URL directly via an
// <a download> / window.open.
export const useRevenueCSVExportURL = (params?: DateRangeParams): string => {
  const { from, to } = resolveDateRange(params);
  const base =
    (import.meta as ImportMeta & { env: Record<string, string> }).env.VITE_API_URL ??
    'https://api.voxire.com/';
  const trimmed = base.endsWith('/') ? base.slice(0, -1) : base;
  const query = new URLSearchParams({ from, to }).toString();
  return `${trimmed}/api/insights/report/v1/reports/revenue/export?${query}`;
};
