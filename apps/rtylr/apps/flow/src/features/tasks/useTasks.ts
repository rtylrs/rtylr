import { unwrap } from '@voxire/api';
import { useQuery, useQueries, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';
import { flowApi, ApiAttachment, ApiTask } from '../../shared/services/flowApi';
import type { ApiTaskDetail } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

const t = (key: string, opts?: Record<string, unknown>): string => i18next.t(key, opts) as string;

export function useTasks(params: {
  projectId?: string;
  status?: ApiTask['status'];
  assigneeId?: string;
  activeOnly?: boolean;
  compact?: boolean;
}) {
  // v2's list endpoint is business-scoped, with project_id, status and
  // assignee_id as its only filters. This used to add
  // `workspaceId: activeWorkspaceId` and gate the query on
  // `projectId || workspaceId`. The workspace layer is gone, so
  // activeWorkspaceId is permanently '', and every caller that passes no
  // projectId was permanently disabled: it rendered an empty state without
  // issuing a request at all. That is triage, archive, the global calendar,
  // Time's My Day and the command bar's task search. `workspaceId` was never
  // a parameter of getTasks either, so it only ever affected this gate.
  return useQuery<ApiTask[]>({
    queryKey: ['tasks', params],
    queryFn: async () => {
      const res = await unwrap(flowApi.getTasks(params));
      return unwrapListResponse<ApiTask>(res, 'tasks');
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 30 * 60 * 1000,
  });
}

export function useMyTasks() {
  return useQuery<ApiTask[]>({
    queryKey: ['my-tasks'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getMyTasks());
      return unwrapListResponse<ApiTask>(res, 'tasks');
    },
  });
}

export function useTaskDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['task', id],
    queryFn: () => unwrap(flowApi.getTaskDetail(id!)),
    enabled: !!id,
  });
}

/**
 * Task details for exactly the ids a list query did not return.
 *
 * The bulk-edit selection is global, but the business-wide task list is a
 * single page (the server caps it), so a task selected on a project board
 * beyond that page is absent from it. Resolving those ids individually
 * completes the selection without the caller having to page the whole
 * business. Shares `['task', id]` with useTaskDetail/useTask.
 */
export function useTaskDetails(ids: string[]) {
  return useQueries({
    queries: ids.map((id) => ({
      queryKey: ['task', id],
      queryFn: async (): Promise<ApiTaskDetail> => unwrap(flowApi.getTaskDetail(id)),
    })),
  });
}

export function useCreateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      // projectId is optional: the backend resolves the caller's auto-created
      // Inbox project when it is omitted (quick-create from My Tasks).
      projectId?: string;
      title: string;
      description?: string;
      status?: ApiTask['status'];
      priority?: ApiTask['priority'];
      assigneeId?: string;
      dueDate?: string;
      effortMinutes?: number;
      estimatedHours?: number;
      labels?: string[];
    }) => {
      return unwrap(flowApi.createTask(data));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.task.createFailed')),
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiTask> }) => {
      return unwrap(flowApi.updateTask(id, data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.task.updateFailed')),
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteTask(id));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.task.deleteFailed')),
  });
}

export function useTaskAttachments(taskId: string | undefined) {
  return useQuery<{ attachments: ApiAttachment[] }>({
    queryKey: ['task-attachments', taskId],
    queryFn: async () => unwrap(flowApi.getAttachments(taskId!)),
    enabled: !!taskId,
  });
}

export function useAddAttachment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      taskId,
      name,
      size,
      mime,
      url,
    }: {
      taskId: string;
      name: string;
      size: number;
      mime: string;
      url?: string;
    }) => unwrap(flowApi.addAttachment(taskId, { name, size, mime, url })),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['task', variables.taskId] });
      qc.invalidateQueries({ queryKey: ['task-attachments', variables.taskId] });
    },
    onError: () => toast.error(t('toasts.attachment.addFailed')),
  });
}

export function useRemoveAttachment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, attachmentId }: { taskId: string; attachmentId: string }) =>
      unwrap(flowApi.removeAttachment(taskId, attachmentId)),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['task', variables.taskId] });
      qc.invalidateQueries({ queryKey: ['task-attachments', variables.taskId] });
    },
    onError: () => toast.error(t('toasts.attachment.removeFailed')),
  });
}

export function useUpdateTaskStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: ApiTask['status'] }) => {
      return unwrap(flowApi.updateTaskStatus(id, status));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.task.statusUpdateFailed')),
  });
}

// Bulk operations: report partial-success accurately and ALWAYS
// invalidate caches on settle, so items that did succeed don't show
// stale state when one item in the batch fails.
type BulkResult = { ok: number; failed: number };

async function bulkSettle<T>(items: T[], op: (item: T) => Promise<unknown>): Promise<BulkResult> {
  const results = await Promise.allSettled(items.map((item) => op(item)));
  let ok = 0;
  let failed = 0;
  for (const r of results) {
    if (r.status === 'fulfilled') ok++;
    else failed++;
  }
  return { ok, failed };
}

function reportBulk(result: BulkResult, successKey: string, failKey: string) {
  if (result.failed === 0) {
    toast.success(t(successKey, { count: result.ok }));
  } else if (result.ok === 0) {
    toast.error(t(failKey));
  } else {
    toast.warning(
      `${result.ok} updated, ${result.failed} failed - refreshing the list to show current state.`
    );
  }
}

export function useBulkUpdateStatus() {
  const qc = useQueryClient();
  return useMutation<BulkResult, unknown, { ids: string[]; status: ApiTask['status'] }>({
    mutationFn: ({ ids, status }) =>
      bulkSettle(ids, (id) => unwrap(flowApi.updateTaskStatus(id, status))),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onSuccess: (res) => reportBulk(res, 'toasts.task.bulkUpdated', 'toasts.task.bulkUpdateFailed'),
    onError: () => toast.error(t('toasts.task.bulkUpdateFailed')),
  });
}

export function useBulkUpdatePriority() {
  const qc = useQueryClient();
  return useMutation<BulkResult, unknown, { ids: string[]; priority: ApiTask['priority'] }>({
    mutationFn: ({ ids, priority }) =>
      bulkSettle(ids, (id) => unwrap(flowApi.updateTask(id, { priority }))),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onSuccess: (res) => reportBulk(res, 'toasts.task.bulkUpdated', 'toasts.task.bulkUpdateFailed'),
    onError: () => toast.error(t('toasts.task.bulkUpdateFailed')),
  });
}

export function useBulkUpdateArchive() {
  const qc = useQueryClient();
  return useMutation<BulkResult, unknown, { ids: string[] }>({
    mutationFn: ({ ids }) =>
      bulkSettle(ids, (id) => unwrap(flowApi.updateTask(id, { archived: true }))),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onSuccess: (res) => reportBulk(res, 'toasts.task.bulkUpdated', 'toasts.task.bulkUpdateFailed'),
    onError: () => toast.error(t('toasts.task.bulkUpdateFailed')),
  });
}

export function useBulkDelete() {
  const qc = useQueryClient();
  return useMutation<BulkResult, unknown, string[]>({
    mutationFn: (ids) => bulkSettle(ids, (id) => unwrap(flowApi.deleteTask(id))),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onSuccess: (res) => reportBulk(res, 'toasts.task.bulkDeleted', 'toasts.task.bulkDeleteFailed'),
    onError: () => toast.error(t('toasts.task.bulkDeleteFailed')),
  });
}

export function useAddComment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { taskId: string; content: string }) => {
      return unwrap(flowApi.addComment(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.comment.postFailed')),
  });
}

export function useDeleteComment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteComment(id));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.comment.deleteFailed')),
  });
}

export function useAddSubtask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, title }: { taskId: string; title: string }) => {
      return unwrap(flowApi.createSubtask(taskId, { title })).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.subtask.addFailed')),
  });
}

export function useToggleSubtask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, completed }: { id: string; completed: boolean }) =>
      unwrap(flowApi.updateTask(id, { status: completed ? 'todo' : 'done' })).then(() => {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.subtask.updateFailed')),
  });
}
