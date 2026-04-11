import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useInsightsSettings() {
  return useQuery({
    queryKey: ['insights', 'settings'],
    queryFn: () => insightsService.getInsightsSettings().then((r) => r.data),
  });
}

export function useUpdateInsightsSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => insightsService.updateInsightsSettings(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'settings'] });
    },
  });
}

export function useScheduledReports() {
  return useQuery({
    queryKey: ['insights', 'scheduled-reports'],
    queryFn: () => insightsService.getScheduledReports().then((r) => r.data),
  });
}

export function useCreateScheduledReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => insightsService.createScheduledReport(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'scheduled-reports'] });
    },
  });
}

export function useDeleteScheduledReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => insightsService.deleteScheduledReport(id).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'scheduled-reports'] });
    },
  });
}
