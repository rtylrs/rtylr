import type { SavedView } from '../stores/flowStore';
import { EMPTY_TASK_FILTERS, parseTaskFilters, type TaskFilters } from './taskFilters';

type ViewFilters = SavedView['filters'];

/**
 * A saved view's filters as the shared filter model.
 *
 * Views saved before multi-select hold one `assigneeId` and one `projectId`;
 * those read as one-value lists, so an old view filters exactly as it did.
 * Statuses on a view are the built-in status names, not a project's columns.
 */
export function viewFiltersToTaskFilters(filters: ViewFilters | undefined): TaskFilters {
  if (!filters) return EMPTY_TASK_FILTERS;
  const parsed = parseTaskFilters({
    assignees: filters.assignees ?? (filters.assigneeId ? [filters.assigneeId] : []),
    projects: filters.projects ?? (filters.projectId ? [filters.projectId] : []),
    priorities: filters.priority,
    statuses: filters.status,
    labels: filters.labels,
    milestones: filters.milestones,
    cycles: filters.cycles,
    due: filters.due,
    dueFrom: filters.dueFrom,
    dueTo: filters.dueTo,
  });
  // A view does not save a search; the box on the Views page is a quick
  // look within the view, not part of it.
  return { ...parsed, query: '' };
}

/**
 * The shared model as a view's stored filters. The old single-value keys are
 * written as undefined so a view edited here is not read back through them.
 */
export function taskFiltersToViewFilters(filters: TaskFilters): ViewFilters {
  return {
    status: filters.statuses as ViewFilters['status'],
    priority: filters.priorities,
    labels: filters.labels,
    assignees: filters.assignees,
    projects: filters.projects,
    milestones: filters.milestones,
    cycles: filters.cycles,
    due: filters.due,
    dueFrom: filters.dueFrom,
    dueTo: filters.dueTo,
    assigneeId: undefined,
    projectId: undefined,
  };
}
