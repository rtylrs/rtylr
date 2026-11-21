import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { flowApi, type ApiTask } from '../../services/flowApi';

/** A task key such as "T-42", "t42" or a bare "42". */
const TASK_KEY = /^t?-?(\d+)$/i;

/** The task id a query names by key, or undefined when it is not a key. */
export function taskIdFromKey(query: string): string | undefined {
  return TASK_KEY.exec(query.trim())?.[1];
}

async function searchTasks(search: string): Promise<ApiTask[]> {
  // Full-text search matches titles and descriptions, never keys, so a key is
  // looked up directly. The detail endpoint applies the same read check.
  const keyId = taskIdFromKey(search);
  if (keyId) {
    try {
      const detail = await unwrap(flowApi.getTaskDetail(keyId));
      return detail.task ? [detail.task] : [];
    } catch {
      return [];
    }
  }
  const res = await unwrap(search.length >= 3 ? flowApi.searchTasks(search) : flowApi.getTasks({}));
  return res.tasks ?? [];
}

/**
 * Tasks matching what someone is typing, by key or by title. The query is
 * debounced; under three characters it falls back to the caller's task list,
 * which the caller filters as they type. Shared by the command bar and the
 * task relation picker.
 */
export function useTaskSearch(query: string, enabled: boolean) {
  const [search, setSearch] = useState(query.trim());
  useEffect(() => {
    const timer = setTimeout(() => setSearch(query.trim()), 250);
    return () => clearTimeout(timer);
  }, [query]);
  return useQuery({
    queryKey: ['tasks', 'search', search],
    queryFn: () => searchTasks(search),
    enabled,
  });
}
