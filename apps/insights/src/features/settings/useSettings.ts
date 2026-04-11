import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';

export function useInsightsSettings() {
  return useQuery({
    queryKey: ['insights', 'settings'],
    queryFn: () => unwrap(insightsService.getInsightsSettings()),
  });
}

export function useUpdateInsightsSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => unwrap(insightsService.updateInsightsSettings(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'settings'] });
    },
  });
}

export function useScheduledReports() {
  return useQuery({
    queryKey: ['insights', 'scheduled-reports'],
    queryFn: () => unwrap(insightsService.getScheduledReports()),
  });
}

export function useCreateScheduledReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => unwrap(insightsService.createScheduledReport(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'scheduled-reports'] });
    },
  });
}

export function useDeleteScheduledReport() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(insightsService.deleteScheduledReport(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['insights', 'scheduled-reports'] });
    },
  });
}
