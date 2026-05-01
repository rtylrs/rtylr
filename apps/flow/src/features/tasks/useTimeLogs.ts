import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiTimeLog } from '../../shared/services/flowApi';

const t = (key: string): string => i18next.t(key) as string;

export interface TimeLogsData {
  logs: ApiTimeLog[];
  estimated_hours: number | null;
  actual_hours: number | null;
}

export function useTimeLogs(taskId: string) {
  return useQuery<TimeLogsData>({
    queryKey: ['time-logs', taskId],
    queryFn: () => unwrap(flowApi.getTimeLogs(taskId)),
    enabled: !!taskId,
  });
}

export function useLogTime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { task_id: number; duration_minutes: number; description?: string }) =>
      unwrap(flowApi.logTime(data)),
    onSuccess: (_data, variables) => {
      const tid = String(variables.task_id);
      qc.invalidateQueries({ queryKey: ['time-logs', tid] });
      qc.invalidateQueries({ queryKey: ['task-detail', tid] });
      qc.invalidateQueries({ queryKey: ['task'] });
      toast.success(t('toasts.time.logged'));
    },
    onError: () => toast.error(t('toasts.time.logFailed')),
  });
}

export function useDeleteTimeLog() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id }: { id: number; task_id: number }) => unwrap(flowApi.deleteTimeLog(id)),
    onSuccess: (_data, variables) => {
      const tid = String(variables.task_id);
      qc.invalidateQueries({ queryKey: ['time-logs', tid] });
      qc.invalidateQueries({ queryKey: ['task-detail', tid] });
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.time.deleteFailed')),
  });
}
