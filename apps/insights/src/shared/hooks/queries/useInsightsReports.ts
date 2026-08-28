import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import {
  useMutation,
  useQuery,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ReportServiceClient } from '@voxire/voxie/insights/report/v1/client';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type { ReportTemplatesResponse, GenerateReportResult } from './types';

const reportClient = createOneKitClient(ReportServiceClient, 'insights');

export const useReportTemplates = (): UseQueryResult<ReportTemplatesResponse> => {
  return useQuery<ReportTemplatesResponse>({
    queryKey: ['insights-report-templates'],
    queryFn: async (): Promise<ReportTemplatesResponse> => {
      const res = await reportClient.listReportTemplates({
        page: 1,
        limit: 50,
        sortBy: '',
        sortDir: '',
        type: '',
      });
      return {
        templates: (res.templates ?? []).map((tpl) => ({
          id: tpl.id ?? '',
          name: tpl.name ?? '',
          description: tpl.description ?? '',
        })),
      };
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
      const { from, to } = resolveDateRange(variables);
      const res = await reportClient.generateRevenueReport({
        fromDate: from,
        toDate: to,
        format: variables.format ?? 'pdf',
      });
      return {
        reportId: res.reportId ?? '',
        url: res.url ?? null,
        status: res.status ?? '',
      };
    },
    onSuccess: (): void => {
      toast.success(t('insights.reports.generateQueued'));
    },
    onError: (): void => {
      toast.error(t('insights.errors.generic'));
    },
  });
};

// The revenue CSV export RPC returns a JSON envelope (filename/contentType/csv)
// rather than streaming text/csv, so the download is built client-side from a
// Blob instead of anchoring directly at a REST URL.
export const useExportRevenueReport = (): UseMutationResult<void, unknown, DateRangeParams> => {
  const { t } = useTranslation();
  return useMutation<void, unknown, DateRangeParams>({
    mutationFn: async (params: DateRangeParams): Promise<void> => {
      const { from, to } = resolveDateRange(params);
      const res = await reportClient.exportRevenueReport({ fromDate: from, toDate: to });
      const blob = new Blob([res.csv ?? ''], { type: res.contentType || 'text/csv' });
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = res.filename || 'revenue-report.csv';
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      URL.revokeObjectURL(url);
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('insights.errors.generic')));
    },
  });
};
