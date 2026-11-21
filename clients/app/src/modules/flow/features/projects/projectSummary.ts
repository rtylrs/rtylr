import type { Task } from '../../shared/stores/flowStore';
import type { ApiCycle, ApiStatus } from '../../shared/services/flowApi';
import {
  EMPTY_TASK_FILTERS,
  PRIORITIES,
  UNASSIGNED,
  matchesTaskFilters,
  type FilterContext,
  type TaskFilters,
} from '../../shared/filters/taskFilters';
import { isClosedTask } from '../../shared/filters/useShowCompleted';
import { isClosedBucket, statusIdOf } from './statusUtils';

/**
 * A number on the project dashboard and the filter that opens exactly those
 * tasks. The count is computed with the same matcher the board and list use,
 * so the number on the card is the number of rows the link shows.
 */
export interface SummaryMetric {
  count: number;
  filters: TaskFilters;
}

export interface WorkloadRow {
  /** User id, UNASSIGNED, or `name:<name>` for a named row with no id. */
  key: string;
  /** Empty when unassigned. */
  name: string;
  /** Every task on the row, open or closed: the share bar. */
  total: number;
  open: number;
  /** Estimated hours on the open tasks. */
  openEstimateHours: number;
  /** This person's open tasks; null when there is no id to filter by. */
  filters: TaskFilters | null;
}

export interface CycleGap extends SummaryMetric {
  cycleId: string;
  cycleName: string;
}

export interface HoursSummary {
  /** Estimated hours over every task that is not cancelled. */
  estimated: number;
  /** Hours logged on those same estimated tasks, so the two compare like for like. */
  loggedOnEstimated: number;
  /** Estimated hours still on open tasks. */
  openEstimated: number;
  /** Open tasks with no estimate. */
  openUnestimated: number;
}

export interface ProjectSummaryData {
  recent: {
    created: SummaryMetric;
    completed: SummaryMetric;
    updated: SummaryMetric;
    dueSoon: SummaryMetric;
  };
  attention: {
    overdue: SummaryMetric;
    /** One per active cycle: its open tasks with nobody on them. */
    unassignedInCycles: CycleGap[];
  };
  /** Open tasks by priority, most urgent first. */
  priorities: Array<{ priority: Task['priority']; metric: SummaryMetric }>;
  workload: WorkloadRow[];
  hours: HoursSummary;
  /** The filter behind each built-in status in the breakdown bar, where the project has one. */
  statusFilters: Partial<Record<Task['status'], TaskFilters>>;
}

export interface SummaryInput {
  /** Not archived. */
  tasks: readonly Task[];
  /**
   * Hours logged per task id: the task's actual_hours, which is every log on
   * it by anyone, so it is a property of the task and not of its assignee.
   */
  loggedHours: ReadonlyMap<string, number>;
  /** The project's statuses (its board columns). */
  statuses: readonly ApiStatus[];
  cycles: readonly ApiCycle[];
  /** Local today, `YYYY-MM-DD`. */
  today: string;
}

const hours = (value: number | undefined | null): number =>
  typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : 0;

export function computeProjectSummary(input: SummaryInput): ProjectSummaryData {
  const { tasks, loggedHours, statuses, cycles, today } = input;
  const ctx: FilterContext = { today, statusOf: (task) => statusIdOf(task) ?? '' };

  // "Open" is said through the status field so the link carries it. A project
  // whose statuses have not loaded falls back to the built-in status for the
  // count, and its link then shows closed tasks too.
  const openStatusIds = statuses.filter((s) => !isClosedBucket(s.bucket)).map((s) => String(s.id));
  const isOpen = (task: Task): boolean => !isClosedTask(task);
  const withOpen = (filters: Partial<TaskFilters>): TaskFilters => ({
    ...EMPTY_TASK_FILTERS,
    ...filters,
    ...(openStatusIds.length > 0 ? { statuses: openStatusIds } : {}),
  });
  const metric = (filters: TaskFilters, openOnly = false): SummaryMetric => ({
    filters,
    count: tasks.filter(
      (task) =>
        matchesTaskFilters(task, filters, ctx) &&
        (!openOnly || openStatusIds.length > 0 || isOpen(task))
    ).length,
  });
  const plain = (filters: Partial<TaskFilters>): TaskFilters => ({
    ...EMPTY_TASK_FILTERS,
    ...filters,
  });

  const recent = {
    created: metric(plain({ changed: ['created_7d'] })),
    completed: metric(plain({ changed: ['completed_7d'] })),
    updated: metric(plain({ changed: ['updated_7d'] })),
    dueSoon: metric(withOpen({ due: ['next_7'] }), true),
  };

  const unassignedInCycles = cycles
    .filter((c) => c.status === 'active')
    .map((c) => {
      const cycleId = String(c.id);
      return {
        cycleId,
        cycleName: c.name,
        ...metric(withOpen({ assignees: [UNASSIGNED], cycles: [cycleId] }), true),
      };
    });

  const attention = {
    overdue: metric(withOpen({ due: ['overdue'] }), true),
    unassignedInCycles,
  };

  const priorities = PRIORITIES.map((priority) => ({
    priority,
    metric: metric(withOpen({ priorities: [priority] }), true),
  }));

  const rows = new Map<string, WorkloadRow>();
  for (const task of tasks) {
    const name = task.assignee || '';
    const key = task.assigneeId ? task.assigneeId : name ? `name:${name}` : UNASSIGNED;
    const row = rows.get(key) ?? {
      key,
      name,
      total: 0,
      open: 0,
      openEstimateHours: 0,
      filters: key.startsWith('name:') ? null : withOpen({ assignees: [key] }),
    };
    row.total += 1;
    if (isOpen(task)) row.openEstimateHours += hours(task.estimate);
    rows.set(key, row);
  }
  for (const row of rows.values()) {
    row.open = row.filters
      ? metric(row.filters, true).count
      : tasks.filter((t) => isOpen(t) && !t.assigneeId && (t.assignee || '') === row.name).length;
  }
  // People first, busiest first; the Unassigned row closes the list.
  const workload = Array.from(rows.values()).sort(
    (a, b) =>
      Number(a.key === UNASSIGNED) - Number(b.key === UNASSIGNED) ||
      b.open - a.open ||
      b.total - a.total ||
      a.name.localeCompare(b.name)
  );

  let estimated = 0;
  let loggedOnEstimated = 0;
  let openEstimated = 0;
  let openUnestimated = 0;
  for (const task of tasks) {
    const estimate = hours(task.estimate);
    if (task.status !== 'cancelled' && estimate > 0) {
      estimated += estimate;
      loggedOnEstimated += hours(loggedHours.get(task.id));
    }
    if (isOpen(task)) {
      if (estimate > 0) openEstimated += estimate;
      else openUnestimated += 1;
    }
  }

  const statusFilters: Partial<Record<Task['status'], TaskFilters>> = {};
  for (const s of statuses) {
    const legacy = s.legacyStatus as Task['status'];
    if (!legacy) continue;
    const current = statusFilters[legacy];
    statusFilters[legacy] = plain({ statuses: [...(current?.statuses ?? []), String(s.id)] });
  }

  return {
    recent,
    attention,
    priorities,
    workload,
    hours: { estimated, loggedOnEstimated, openEstimated, openUnestimated },
    statusFilters,
  };
}
