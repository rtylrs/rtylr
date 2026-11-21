import type { Task } from '../../../shared/stores/flowStore';
import type { ApiStatus } from '../../../shared/services/flowApi';
import { PRIORITY_LABEL_KEY } from '../taskGroups';
import { displayStatusName, sortStatuses } from '../statusUtils';
import { listColumn, resolveTaskStatus, type ListColumnId } from './listColumns';
import { buildCsv } from '../../../../../shared/lib/csv';

export interface CsvContext {
  t: (key: string) => string;
  statuses: readonly ApiStatus[];
  labelNames: ReadonlyMap<string, string>;
}

const dateOnly = (value: string | undefined): string => (value ? value.slice(0, 10) : '');

function cellText(
  task: Task,
  column: ListColumnId,
  ctx: CsvContext,
  ordered: readonly ApiStatus[]
): string | number {
  switch (column) {
    case 'priority':
      return ctx.t(PRIORITY_LABEL_KEY[task.priority]);
    case 'title':
      return task.title;
    case 'status': {
      const status = resolveTaskStatus(task, ordered);
      return status ? displayStatusName(status.name, ctx.t) : (task.statusName ?? task.status);
    }
    case 'assignee':
      return task.assignee;
    case 'startDate':
      return dateOnly(task.startDate);
    case 'dueDate':
      return dateOnly(task.dueDate);
    case 'estimate':
      return task.estimate != null && task.estimate > 0 ? task.estimate : '';
    case 'labels':
      return (task.labelIds ?? [])
        .map((id) => ctx.labelNames.get(String(id)))
        .filter((name): name is string => !!name)
        .join('; ');
    case 'key':
      return task.identifier;
    case 'created':
      return dateOnly(task.createdAt);
  }
}

/**
 * The rows as shown: the columns the user picked, in their order, led by the
 * task key so every exported row can be traced back to its task.
 */
export function listTasksToCsv(
  tasks: readonly Task[],
  columns: readonly ListColumnId[],
  ctx: CsvContext
): string {
  const exported: ListColumnId[] = ['key', ...columns.filter((column) => column !== 'key')];
  const ordered = sortStatuses(ctx.statuses);
  const header = exported.map((column) => ctx.t(listColumn(column).labelKey));
  const body = tasks.map((task) => exported.map((column) => cellText(task, column, ctx, ordered)));
  return buildCsv([header, ...body]);
}

/** A file name safe on every desktop OS, e.g. `Website 2026-11-18.csv`. */
export function csvFileName(projectName: string, today: string): string {
  const safe = Array.from(projectName, (ch) =>
    ch.charCodeAt(0) < 32 || '\\/:*?"<>|'.includes(ch) ? ' ' : ch
  )
    .join('')
    .replace(/\s+/g, ' ')
    .trim();
  return `${safe || 'tasks'} ${today}.csv`;
}
