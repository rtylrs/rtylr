import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { Task } from '../../shared/stores/flowStore';
import { useTaskFilters, type TaskFilterState } from '../../shared/filters/useTaskFilters';
import { matchesTaskFilters } from '../../shared/filters/taskFilters';
import {
  labelField,
  namedField,
  priorityField,
  type FilterFieldConfig,
} from '../../shared/filters/filterFields';
import { statusLabelKey } from '../../shared/lib/statusLabels';
import { todayIso } from '../../shared/lib/dateUtils';
import { useLabels } from './useLabels';

const STATUSES: Task['status'][] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
  'cancelled',
];

const NO_EXTRA_FIELDS: FilterFieldConfig[] = [];

export interface MyTaskFilters {
  state: TaskFilterState;
  fields: FilterFieldConfig[];
  /** Ids of the tasks that pass, or null when nothing is filtering. */
  matchingIds: ReadonlySet<string> | null;
}

/**
 * Filters for a cross-project task list (My tasks, and Team tasks with its own
 * scope). Statuses here are the built-in ones, because each project names its
 * own columns and there is no shared set to offer across projects.
 */
export function useCrossProjectTaskFilters(
  scope: string,
  tasks: Task[],
  extraFields: FilterFieldConfig[] = NO_EXTRA_FIELDS
): MyTaskFilters {
  const { t } = useTranslation();
  const state = useTaskFilters(scope);
  const { filters, active } = state;
  const labelsQuery = useLabels();

  const labelNames = useMemo(
    () => new Map((labelsQuery.data ?? []).map((l) => [String(l.id), l.name])),
    [labelsQuery.data]
  );

  const projects = useMemo(() => {
    const byId = new Map<string, string>();
    for (const task of tasks) {
      if (task.projectId && !byId.has(task.projectId)) {
        byId.set(task.projectId, task.projectName || task.projectId);
      }
    }
    return [...byId]
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [tasks]);

  const fields = useMemo(
    (): FilterFieldConfig[] => [
      ...extraFields,
      namedField('projects', t('flow.filters.fields.project'), projects),
      priorityField(t),
      labelField(labelsQuery.data ?? [], t),
      namedField(
        'statuses',
        t('flow.filters.fields.status'),
        STATUSES.map((s) => ({ id: s, name: t(statusLabelKey(s)) }))
      ),
    ],
    [extraFields, projects, labelsQuery.data, t]
  );

  const matchingIds = useMemo(() => {
    if (!active) return null;
    const ctx = { today: todayIso(), labelNames };
    return new Set(tasks.filter((task) => matchesTaskFilters(task, filters, ctx)).map((x) => x.id));
  }, [active, filters, tasks, labelNames]);

  return { state, fields, matchingIds };
}

/** My tasks: the signed-in user's tasks across every project. */
export function useMyTaskFilters(tasks: Task[]): MyTaskFilters {
  return useCrossProjectTaskFilters('my-tasks', tasks);
}
