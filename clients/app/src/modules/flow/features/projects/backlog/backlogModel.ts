/**
 * The Backlog screen's data model: which tasks go in which section, in what
 * order, with which totals, and what a drop means to the server.
 *
 * Kept apart from the React wiring, like `boardDrag.ts` and `boardOrder.ts`,
 * so the rules can be tested without a DOM.
 */
import type { Task } from '../../../shared/stores/flowStore';
import type { ApiCycle, ApiStatus } from '../../../shared/services/flowApi';
import { statusIdOf, type StatusBucket } from '../statusUtils';

/** The id the unplanned section uses where a cycle id would go. */
export const BACKLOG_SECTION_ID = '';

export interface BacklogTotals {
  todo: number;
  active: number;
  done: number;
  /** Summed effort estimate, in minutes. */
  estimateMinutes: number;
}

export interface BacklogSection {
  /** The cycle id, or BACKLOG_SECTION_ID for unplanned work. */
  id: string;
  cycle?: ApiCycle;
  tasks: Task[];
  totals: BacklogTotals;
}

const LEGACY_BUCKETS: Record<string, StatusBucket> = {
  backlog: 'backlog',
  todo: 'backlog',
  in_progress: 'active',
  in_review: 'active',
  done: 'done',
  cancelled: 'cancelled',
};

/**
 * A task's status bucket: its custom status's bucket when it has one in the
 * project, else the bucket of its built-in status. The same rule the server
 * uses for cycle totals.
 */
export function taskBucket(task: Task, statusesById: ReadonlyMap<string, ApiStatus>): StatusBucket {
  const statusId = statusIdOf(task);
  const custom = statusId ? statusesById.get(statusId) : undefined;
  return custom?.bucket ?? LEGACY_BUCKETS[task.status] ?? 'backlog';
}

/**
 * Tasks in plan order: ranked ones by rank, then the never-placed ones oldest
 * first, which is how the server orders the unplanned backlog.
 */
export function orderByPlanRank(tasks: readonly Task[]): Task[] {
  return [...tasks].sort((a, b) => {
    const ra = a.planRank;
    const rb = b.planRank;
    if (ra !== undefined && rb !== undefined && ra !== rb) return ra - rb;
    if (ra !== undefined && rb === undefined) return -1;
    if (ra === undefined && rb !== undefined) return 1;
    return (a.createdAt ?? '').localeCompare(b.createdAt ?? '');
  });
}

export function sectionTotals(
  tasks: readonly Task[],
  statusesById: ReadonlyMap<string, ApiStatus>
): BacklogTotals {
  const totals: BacklogTotals = { todo: 0, active: 0, done: 0, estimateMinutes: 0 };
  for (const task of tasks) {
    const bucket = taskBucket(task, statusesById);
    // Cancelled work is out of scope, as in the server's cycle totals.
    if (bucket === 'cancelled') continue;
    if (bucket === 'active') totals.active += 1;
    else if (bucket === 'done') totals.done += 1;
    else totals.todo += 1;
    totals.estimateMinutes += task.effortMinutes ?? 0;
  }
  return totals;
}

/**
 * The sections in screen order: the active cycle, then upcoming cycles by start
 * date, then unplanned work. Completed cycles are history and stay on the
 * Cycles page. The unplanned section leaves out finished tasks, which nobody
 * plans; a cycle keeps them so its done count is visible.
 */
export function buildBacklogSections(
  cycles: readonly ApiCycle[],
  tasks: readonly Task[],
  statusesById: ReadonlyMap<string, ApiStatus>
): BacklogSection[] {
  const open = cycles.filter((c) => c.status === 'active' || c.status === 'upcoming');
  const ordered = [...open].sort((a, b) => {
    if (a.status !== b.status) return a.status === 'active' ? -1 : 1;
    return (a.startDate ?? '').localeCompare(b.startDate ?? '');
  });
  const openIds = new Set(ordered.map((c) => c.id));
  const byCycle = new Map<string, Task[]>();
  const unplanned: Task[] = [];
  for (const task of tasks) {
    if (task.archived) continue;
    if (task.cycleId && openIds.has(task.cycleId)) {
      const list = byCycle.get(task.cycleId) ?? [];
      list.push(task);
      byCycle.set(task.cycleId, list);
    } else if (!task.cycleId) {
      const bucket = taskBucket(task, statusesById);
      if (bucket !== 'done' && bucket !== 'cancelled') unplanned.push(task);
    }
  }
  const sections: BacklogSection[] = ordered.map((cycle) => {
    const list = orderByPlanRank(byCycle.get(cycle.id) ?? []);
    return { id: cycle.id, cycle, tasks: list, totals: sectionTotals(list, statusesById) };
  });
  const backlog = orderByPlanRank(unplanned);
  sections.push({
    id: BACKLOG_SECTION_ID,
    tasks: backlog,
    totals: sectionTotals(backlog, statusesById),
  });
  return sections;
}

export interface PlannedMove {
  taskId: string;
  cycleId: string;
  beforeTaskId?: string;
  afterTaskId?: string;
  optimisticRank: number;
}

const RANK_GAP = 1024;

/**
 * What dropping `taskId` at `index` of a section (dragged task excluded from
 * the count) asks the server for. The anchor is the task the drop lands
 * before, or after when it lands at the end. The optimistic rank only has to
 * sort correctly until the server answers.
 */
export function planMove(section: BacklogSection, taskId: string, index: number): PlannedMove {
  const others = section.tasks.filter((t) => t.id !== taskId);
  const at = Math.max(0, Math.min(index, others.length));
  const next = others[at];
  const prev = at > 0 ? others[at - 1] : undefined;
  const prevRank = prev?.planRank;
  const nextRank = next?.planRank;
  let optimisticRank = 0;
  if (prevRank !== undefined && nextRank !== undefined) optimisticRank = (prevRank + nextRank) / 2;
  else if (nextRank !== undefined) optimisticRank = nextRank - RANK_GAP;
  else if (prevRank !== undefined) optimisticRank = prevRank + RANK_GAP;

  if (next) return { taskId, cycleId: section.id, beforeTaskId: next.id, optimisticRank };
  if (prev) return { taskId, cycleId: section.id, afterTaskId: prev.id, optimisticRank };
  return { taskId, cycleId: section.id, optimisticRank };
}

export type CapacityTone = 'ok' | 'warn' | 'over';

/** Amber from 90% of the available time, red once it is exceeded. */
export function capacityTone(committedMinutes: number, availableMinutes: number): CapacityTone {
  if (availableMinutes <= 0) return committedMinutes > 0 ? 'over' : 'ok';
  const load = committedMinutes / availableMinutes;
  if (load > 1) return 'over';
  if (load >= 0.9) return 'warn';
  return 'ok';
}

/** Hours from minutes, at most one decimal, for display. */
export function hoursLabel(minutes: number): string {
  return `${Math.round((minutes / 60) * 10) / 10}h`;
}

/** The weeks a cycle can run for. */
export const CYCLE_LENGTH_WEEKS = [1, 2, 3, 4] as const;
