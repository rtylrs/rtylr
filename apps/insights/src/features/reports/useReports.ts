import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';

export function useReportTemplates() {
  return useQuery({
    queryKey: ['insights', 'report-templates'],
    queryFn: () => unwrap(insightsService.getReportTemplates()),
  });
}

export function useGenerateReport() {
  return useMutation({
    mutationFn: ({ templateId, params }: { templateId: string; params?: Record<string, string> }) =>
      unwrap(insightsService.generateReport(templateId, params)),
  });
}

export function useExportReport() {
  return useMutation({
    mutationFn: ({
      templateId,
      format,
      params,
    }: {
      templateId: string;
      format: 'csv' | 'pdf';
      params?: Record<string, string>;
    }) => unwrap(insightsService.exportReport(templateId, format, params)),
  });
}
