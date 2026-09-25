import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiTask } from '../../shared/services/flowApi';

const MY_TASKS_PAGE_SIZE = 500;

/**
 * The endpoint is paginated, but Today needs the complete eligible set rather
 * than the first page. Merge pages using the server's continuation metadata so
 * an overdue backlog cannot hide a due-today task on page two.
 */
export async function fetchAllMyTasks(): Promise<ApiTask[]> {
  const all: ApiTask[] = [];
  let page = 1;

  for (;;) {
    const res = await unwrap(flowApi.getMyTasks({ page, limit: MY_TASKS_PAGE_SIZE }));
    const rows = Array.isArray(res) ? res : (res.tasks ?? []);
    all.push(...rows);

    const pagination = Array.isArray(res) ? undefined : res.pagination;
    const hasNext =
      pagination?.hasNext ??
      (typeof pagination?.total === 'number' && pagination.total > all.length);
    if (!hasNext || rows.length === 0) return all;

    page = (pagination?.page ?? page) + 1;
  }
}

/** Tasks assigned to the current user (My tasks). */
export function useMyIssues() {
  return useQuery<ApiTask[]>({
    queryKey: ['my-tasks'],
    queryFn: fetchAllMyTasks,
  });
}
