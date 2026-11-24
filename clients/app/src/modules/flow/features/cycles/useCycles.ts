import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type {
  ApiCycle,
  ApiCycleCapacity,
  ApiCycleTaskDecision,
  ApiTask,
} from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

// The wire omits zero counts, so a cycle with nothing done yet arrives without
// `completed` (and an empty one without `scope`).
const withCounts = (cycle: ApiCycle): ApiCycle => ({
  ...cycle,
  scope: cycle.scope ?? 0,
  completed: cycle.completed ?? 0,
});

/**
 * The query behind useCycles, shared so a screen that needs several projects'
 * cycles at once (the Plan calendar) reads and fills the same cache entries.
 */
export function cyclesQueryOptions(projectId?: string) {
  return {
    queryKey: ['cycles', projectId],
    queryFn: async (): Promise<ApiCycle[]> => {
      const res = await unwrap(flowApi.getCycles(projectId!));
      return unwrapListResponse<ApiCycle>(res, 'cycles').map(withCounts);
    },
    enabled: projectId !== undefined && projectId !== '',
  };
}

/**
 * Cycles for one project.
 *
 * v2 requires a project id, so this stays disabled until the caller has one
 * rather than firing a request that would 400.
 */
export function useCycles(projectId?: string): ReturnType<typeof useQuery<ApiCycle[]>> {
  return useQuery<ApiCycle[]>(cyclesQueryOptions(projectId));
}

export function useCreateCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiCycle> & { projectId: string }) => {
      return unwrap(flowApi.createCycle(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
    onError: () => toast.error(t('flowToasts.cycle.createFailed')),
  });
}

export function useUpdateCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiCycle> }) => {
      return unwrap(flowApi.updateCycle(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
    onError: () => toast.error(t('flowToasts.cycle.updateFailed')),
  });
}

/** One cycle, for its detail page. */
export function useCycle(id: string | undefined) {
  return useQuery<ApiCycle>({
    queryKey: ['cycles', 'detail', id],
    queryFn: async () => withCounts((await unwrap(flowApi.getCycle(id!))).cycle),
    enabled: id !== undefined && id !== '',
  });
}

export function useDeleteCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.deleteCycle(id)).then(() => {}),
    onSuccess: () => invalidatePlanning(qc),
    onError: () => toast.error(t('flowToasts.cycle.deleteFailed')),
  });
}

/** Whether a failed request was refused as a conflict (HTTP 409). */
function isConflict(error: unknown): boolean {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    (error as { status?: unknown }).status === 409
  );
}

/**
 * Everything a cycle change can move: the cycle list (totals, carry-over), the
 * tasks it holds, and the task panel's cycle field.
 */
function invalidatePlanning(qc: ReturnType<typeof useQueryClient>) {
  void qc.invalidateQueries({ queryKey: ['cycles'] });
  void qc.invalidateQueries({ queryKey: ['cycle-capacity'] });
  void qc.invalidateQueries({ queryKey: ['tasks'] });
  void qc.invalidateQueries({ queryKey: ['my-tasks'] });
  void qc.invalidateQueries({ queryKey: ['task'] });
}

export function useStartCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, lengthWeeks, goal }: { id: string; lengthWeeks: number; goal?: string }) =>
      unwrap(flowApi.startCycle(id, { lengthWeeks, goal })),
    onSuccess: () => {
      invalidatePlanning(qc);
      toast.success(t('flowToasts.cycle.started'));
    },
    onError: (error) =>
      toast.error(
        isConflict(error) ? t('flowToasts.cycle.alreadyActive') : t('flowToasts.cycle.startFailed')
      ),
  });
}

export function useCompleteCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, decisions }: { id: string; decisions: ApiCycleTaskDecision[] }) =>
      unwrap(flowApi.completeCycle(id, decisions)),
    onSuccess: () => {
      invalidatePlanning(qc);
      toast.success(t('flowToasts.cycle.completed'));
    },
    onError: () => toast.error(t('flowToasts.cycle.completeFailed')),
  });
}

export interface CycleTaskMove {
  taskId: string;
  /** Empty for the unplanned backlog. */
  cycleId: string;
  beforeTaskId?: string;
  afterTaskId?: string;
  /** The rank to show until the server answers. */
  optimisticRank: number;
}

/**
 * Places a task in a cycle or the backlog at a position. The task moves in
 * every cached task list at once, so the drop lands without waiting for the
 * round trip; a failure rolls it back.
 */
export function useMoveCycleTask() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (move: CycleTaskMove) =>
      unwrap(
        flowApi.moveCycleTask({
          taskId: move.taskId,
          cycleId: move.cycleId,
          beforeTaskId: move.beforeTaskId,
          afterTaskId: move.afterTaskId,
        })
      ),
    onMutate: async (move) => {
      await qc.cancelQueries({ queryKey: ['tasks'] });
      const previous = qc.getQueriesData<ApiTask[]>({ queryKey: ['tasks'] });
      qc.setQueriesData<ApiTask[]>({ queryKey: ['tasks'] }, (tasks) =>
        Array.isArray(tasks)
          ? tasks.map((task) =>
              String(task.id) === move.taskId
                ? {
                    ...task,
                    cycleId: move.cycleId || undefined,
                    planRank: String(move.optimisticRank),
                  }
                : task
            )
          : tasks
      );
      return { previous };
    },
    onError: (_error, _move, context) => {
      context?.previous.forEach(([key, data]) => qc.setQueryData(key, data));
      toast.error(t('flowToasts.cycle.moveFailed'));
    },
    onSettled: () => invalidatePlanning(qc),
  });
}

export function useCycleCapacity(cycleId: string | undefined) {
  return useQuery<ApiCycleCapacity>({
    queryKey: ['cycle-capacity', cycleId],
    queryFn: () => unwrap(flowApi.getCycleCapacity(cycleId!)),
    enabled: cycleId !== undefined && cycleId !== '',
    staleTime: 60 * 1000,
  });
}
