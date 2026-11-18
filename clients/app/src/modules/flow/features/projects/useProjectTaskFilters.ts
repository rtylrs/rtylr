import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { Task } from '../../shared/stores/flowStore';
import { useTaskFilters, type TaskFilterState } from '../../shared/filters/useTaskFilters';
import { matchesTaskFilters } from '../../shared/filters/taskFilters';
import {
  assigneeField,
  labelField,
  namedField,
  peopleFrom,
  priorityField,
  type FilterFieldConfig,
  type FilterPerson,
} from '../../shared/filters/filterFields';
import { useCurrentUserId } from '../../shared/lib/currentUser';
import { todayIso } from '../../shared/lib/dateUtils';
import { useProjectMembers } from './useProjectMembers';
import { useProjectStatuses } from './useStatuses';
import { displayStatusName, sortStatuses, statusIdOf } from './statusUtils';
import { useLabels } from '../tasks/useLabels';
import { useCycles } from '../cycles/useCycles';
import { useMilestones } from '../milestones/useMilestones';

export interface ProjectTaskFilters {
  state: TaskFilterState;
  fields: FilterFieldConfig[];
  people: FilterPerson[];
  currentUserId: string | null;
  /** Ids of the tasks that pass, or null when nothing is filtering. */
  matchingIds: ReadonlySet<string> | null;
  /** The tasks a view shows: only matches, unless dimming (then all of them). */
  shownTasks: Task[];
  /** Tasks to fade, while "Dim instead of hide" is on and a filter is active. */
  dimmedIds: ReadonlySet<string> | null;
}

/**
 * Everything a project view needs to filter its tasks: the shared state
 * (URL + remembered per project, so board, list and calendar agree), the
 * popover's fields built from this project's people, labels, statuses,
 * cycles and milestones, and which tasks pass.
 *
 * Statuses are filtered by status id, because a project's columns are its
 * statuses.
 */
export function useProjectTaskFilters(projectId: string, tasks: Task[]): ProjectTaskFilters {
  const { t } = useTranslation();
  const state = useTaskFilters(`project:${projectId}`);
  const { filters, active, dim } = state;
  const currentUserId = useCurrentUserId();
  const membersQuery = useProjectMembers(projectId);
  const labelsQuery = useLabels();
  const statusesQuery = useProjectStatuses(projectId);
  const cyclesQuery = useCycles(projectId);
  const milestonesQuery = useMilestones(projectId);

  const people = useMemo(
    () => peopleFrom(membersQuery.data ?? [], tasks),
    [membersQuery.data, tasks]
  );
  const labelNames = useMemo(
    () => new Map((labelsQuery.data ?? []).map((l) => [String(l.id), l.name])),
    [labelsQuery.data]
  );

  const fields = useMemo((): FilterFieldConfig[] => {
    const out: FilterFieldConfig[] = [
      assigneeField(people, currentUserId, t),
      priorityField(t),
      labelField(labelsQuery.data ?? [], t),
      namedField(
        'statuses',
        t('flow.filters.fields.status'),
        sortStatuses(statusesQuery.data ?? []).map((c) => ({
          id: c.id,
          name: displayStatusName(c.name, t),
          color: c.color,
        }))
      ),
    ];
    const cycles = cyclesQuery.data ?? [];
    if (cycles.length > 0) out.push(namedField('cycles', t('flow.filters.fields.cycle'), cycles));
    const milestones = milestonesQuery.data ?? [];
    if (milestones.length > 0) {
      out.push(namedField('milestones', t('flow.filters.fields.milestone'), milestones));
    }
    return out;
  }, [
    people,
    currentUserId,
    labelsQuery.data,
    statusesQuery.data,
    cyclesQuery.data,
    milestonesQuery.data,
    t,
  ]);

  const matchingIds = useMemo(() => {
    if (!active) return null;
    const ctx = {
      today: todayIso(),
      labelNames,
      statusOf: (task: Task) => statusIdOf(task) ?? '',
    };
    return new Set(tasks.filter((task) => matchesTaskFilters(task, filters, ctx)).map((x) => x.id));
  }, [active, filters, tasks, labelNames]);

  const shownTasks = useMemo(
    () => (matchingIds !== null && !dim ? tasks.filter((task) => matchingIds.has(task.id)) : tasks),
    [matchingIds, dim, tasks]
  );
  const dimmedIds = useMemo(
    () =>
      matchingIds !== null && dim
        ? new Set(tasks.filter((task) => !matchingIds.has(task.id)).map((x) => x.id))
        : null,
    [matchingIds, dim, tasks]
  );

  return { state, fields, people, currentUserId, matchingIds, shownTasks, dimmedIds };
}
