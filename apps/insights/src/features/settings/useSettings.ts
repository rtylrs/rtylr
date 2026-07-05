import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { toast } from 'sonner';
import {
  insightsService,
  InsightsSettingsPayload,
  ScheduledReportPayload,
} from '../../shared/services/insightsService';
import { insightsSettingsSchema, scheduledReportSchema } from '../../shared/schemas/settings';
import { mapInsightsSettingsResponse, mapScheduledReportsResponse } from '../../shared/utils/onekitMappers';

export function useInsightsSettings() {
  return useQuery({
    queryKey: ['insights', 'settings'],
    queryFn: async () => {
      const res = await unwrap(insightsService.getInsightsSettings());
      return insightsSettingsSchema.parse(mapInsightsSettingsResponse(res as Record<string, unknown>));
    },
  });
}

export function useUpdateInsightsSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: InsightsSettingsPayload) =>
      unwrap(insightsService.updateInsightsSettings(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'settings'] });
    },
    onError: () => toast.error('Failed to update settings'),
  });
}

export function useScheduledReports() {
  return useQuery({
    queryKey: ['insights', 'scheduled-reports'],
    queryFn: async () => {
      const res = await unwrap(insightsService.getScheduledReports());
      const mapped = mapScheduledReportsResponse(res as Record<string, unknown>);
      return z.array(scheduledReportSchema).parse(mapped.scheduledReports);
    },
  });
}

export function useCreateScheduledReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: ScheduledReportPayload) =>
      unwrap(insightsService.createScheduledReport(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'scheduled-reports'] });
    },
    onError: () => toast.error('Failed to create scheduled report'),
  });
}

export function useDeleteScheduledReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(insightsService.deleteScheduledReport(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'scheduled-reports'] });
    },
    onError: () => toast.error('Failed to delete scheduled report'),
  });
}
