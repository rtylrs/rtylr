import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useReportTemplates() {
  return useQuery({
    queryKey: ['insights', 'report-templates'],
    queryFn: () => insightsService.getReportTemplates().then((r) => r.data),
  });
}

export function useGenerateReport() {
  return useMutation({
    mutationFn: ({ templateId, params }: { templateId: string; params?: Record<string, string> }) =>
      insightsService.generateReport(templateId, params).then((r) => r.data),
  });
}

export function useExportReport() {
  return useMutation({
    mutationFn: ({ templateId, format, params }: { templateId: string; format: 'csv' | 'pdf'; params?: Record<string, string> }) =>
      insightsService.exportReport(templateId, format, params).then((r) => r.data),
  });
}
