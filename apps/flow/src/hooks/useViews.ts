import { useMemo } from 'react';
import { useFlowStore, SavedView, Task, PRIORITY_ORDER } from '../stores/flowStore';

export function useViews() {
  const { views } = useFlowStore();
  return views;
}

export function useFilteredTasks(view: SavedView | null) {
  const { tasks } = useFlowStore();
  return useMemo(() => {
    if (!view) return tasks;
    let filtered = [...tasks];
    if (view.filters.status && view.filters.status.length) {
      filtered = filtered.filter((t) => view.filters.status!.includes(t.status));
    }
    if (view.filters.priority && view.filters.priority.length) {
      filtered = filtered.filter((t) => view.filters.priority!.includes(t.priority));
    }
    if (view.filters.assigneeId) {
      filtered = filtered.filter((t) => t.assigneeId === view.filters.assigneeId);
    }
    if (view.filters.projectId) {
      filtered = filtered.filter((t) => t.projectId === view.filters.projectId);
    }
    if (view.filters.labels && view.filters.labels.length) {
      filtered = filtered.filter((t) =>
        t.labels?.some((l) => view.filters.labels!.includes(l))
      );
    }
    if (view.sortBy) {
      const dir = view.sortDir === 'desc' ? -1 : 1;
      filtered.sort((a, b) => {
        switch (view.sortBy) {
          case 'priority':
            return (PRIORITY_ORDER[a.priority] - PRIORITY_ORDER[b.priority]) * dir;
          case 'dueDate':
            return a.dueDate.localeCompare(b.dueDate) * dir;
          case 'updated':
            return a.updatedAt.localeCompare(b.updatedAt) * dir;
          case 'created':
            return a.createdAt.localeCompare(b.createdAt) * dir;
          default:
            return 0;
        }
      });
    }
    return filtered;
  }, [tasks, view]);
}

export function groupTasks(
  tasks: Task[],
  groupBy: SavedView['groupBy']
): { key: string; label: string; tasks: Task[] }[] {
  if (!groupBy) return [{ key: 'all', label: 'All', tasks }];
  const map: Record<string, Task[]> = {};
  tasks.forEach((t) => {
    const key = t[groupBy as keyof Task] as string;
    if (!map[key]) map[key] = [];
    map[key].push(t);
  });
  return Object.entries(map).map(([key, items]) => ({
    key,
    label: String(key).replace('_', ' '),
    tasks: items,
  }));
}
