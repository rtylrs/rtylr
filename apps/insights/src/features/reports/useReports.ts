import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import { reportTemplateSchema } from '../../shared/schemas/reports';

export function useReportTemplates() {
  return useQuery({
    queryKey: ['insights', 'report-templates'],
    queryFn: async () => {
      const res = await unwrap(insightsService.getReportTemplates());
      const raw = (res as Record<string, unknown>)?.templates ?? res;
      return z.array(reportTemplateSchema).parse(Array.isArray(raw) ? raw : []);
    },
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
