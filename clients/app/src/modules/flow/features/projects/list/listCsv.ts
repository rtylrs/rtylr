import type { Task } from '../../../shared/stores/flowStore';
import type { ApiStatus } from '../../../shared/services/flowApi';
import { PRIORITY_LABEL_KEY } from '../taskGroups';
import { displayStatusName, sortStatuses } from '../statusUtils';
import { listColumn, resolveTaskStatus, type ListColumnId } from './listColumns';

/**
 * Characters that make a spreadsheet read a cell as a formula (or, for tab
 * and carriage return, that some of them strip before doing so). A task
 * title such as `=HYPERLINK(...)` must arrive as text, not run.
 */
const FORMULA_START = /^[=+\-@\t\r]/;

/**
 * One CSV field: always quoted, quotes doubled, and a value that would start
 * a formula prefixed with an apostrophe so it stays text.
 */
export function csvField(value: string | number | null | undefined): string {
  let text = value == null ? '' : String(value);
  if (FORMULA_START.test(text)) text = `'${text}`;
  return `"${text.replace(/"/g, '""')}"`;
}

/**
 * A whole document. The byte order mark makes Excel read it as UTF-8, which
 * Arabic and accented names need; CRLF is what RFC 4180 asks for.
 */
export function buildCsv(rows: ReadonlyArray<ReadonlyArray<string | number | null>>): string {
  return `\uFEFF${rows.map((row) => row.map(csvField).join(',')).join('\r\n')}\r\n`;
}

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

/** Hands the CSV to the browser as a download. Nothing leaves the device. */
export function downloadCsv(fileName: string, csv: string): void {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName;
  link.rel = 'noopener';
  document.body.appendChild(link);
  link.click();
  link.remove();
  // Revoked after the click has been handled, not during it.
  setTimeout(() => URL.revokeObjectURL(url), 0);
}
