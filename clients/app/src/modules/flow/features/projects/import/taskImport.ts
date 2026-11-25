import { parseCsv } from '../../../../../shared/lib/csv';

/**
 * CSV task import (T-343): reads a sheet into rows the import endpoint takes.
 * Names in the file (assignee email, section) are matched here against the
 * project's people and statuses, so problems show in the preview before
 * anything is sent.
 */

export type ImportColumn =
  'title' | 'description' | 'assignee' | 'due' | 'priority' | 'effort' | 'section' | 'subtasks';

// Header spellings people use for each column, compared lower-cased and
// without spaces, dashes or underscores.
const HEADER_ALIASES: Record<ImportColumn, string[]> = {
  title: ['title', 'name', 'task', 'taskname', 'summary'],
  description: ['description', 'details', 'notes', 'body'],
  assignee: ['assignee', 'assigneeemail', 'email', 'assignedto', 'owner'],
  due: ['due', 'duedate', 'deadline'],
  priority: ['priority'],
  effort: ['effort', 'estimate', 'effortminutes', 'minutes', 'hours'],
  section: ['section', 'status', 'column', 'stage'],
  subtasks: ['subtasks', 'checklist', 'subtask'],
};

export const SAMPLE_HEADERS = [
  'Title',
  'Description',
  'Assignee email',
  'Due date',
  'Priority',
  'Effort',
  'Section',
  'Subtasks',
];

export const MAX_IMPORT_ROWS = 500;

const PRIORITIES = new Set(['none', 'low', 'medium', 'high', 'urgent']);

export type ImportIssue =
  | { kind: 'noTitle' }
  | { kind: 'unknownAssignee'; value: string }
  | { kind: 'badDate'; value: string }
  | { kind: 'badPriority'; value: string }
  | { kind: 'badEffort'; value: string }
  | { kind: 'unknownSection'; value: string };

export interface ImportPreviewRow {
  /** 1-based, counting the first data row as 1. */
  line: number;
  title: string;
  description?: string;
  assigneeId?: string;
  assigneeLabel?: string;
  dueDate?: string;
  priority?: string;
  effortMinutes?: number;
  sectionId?: string;
  sectionLabel?: string;
  subtasks: string[];
  issues: ImportIssue[];
}

export interface ImportPreview {
  /** Columns the header row was understood as; unknown headers are ignored. */
  columns: ImportColumn[];
  rows: ImportPreviewRow[];
  /** True when the file has no recognisable Title column. */
  missingTitle: boolean;
  /** True when the file had more rows than MAX_IMPORT_ROWS; the rest were dropped. */
  truncated: boolean;
}

export interface ImportContext {
  people: ReadonlyArray<{ id: string; name: string; email?: string }>;
  statuses: ReadonlyArray<{ id: string; name: string }>;
}

const normalizeHeader = (value: string): string => value.toLowerCase().replace(/[\s_\-.]/g, '');

function columnFor(header: string): ImportColumn | undefined {
  const key = normalizeHeader(header);
  return (Object.keys(HEADER_ALIASES) as ImportColumn[]).find((column) =>
    HEADER_ALIASES[column].includes(key)
  );
}

/**
 * Effort as minutes: "90", "1h", "30m", "1.5h", "1h 30m". A bare number is
 * minutes, except in a column headed "hours".
 */
export function parseEffort(value: string, bareIsHours = false): number | null {
  const text = value.trim().toLowerCase();
  if (text === '') return null;
  if (/^\d+(\.\d+)?$/.test(text)) {
    const n = Number(text);
    return Math.round(bareIsHours ? n * 60 : n);
  }
  const match = /^(?:(\d+(?:\.\d+)?)\s*h(?:ours?|rs?)?)?\s*(?:(\d+)\s*m(?:in(?:utes?|s)?)?)?$/.exec(
    text
  );
  if (!match || (!match[1] && !match[2])) return null;
  return Math.round(Number(match[1] ?? 0) * 60 + Number(match[2] ?? 0));
}

/** YYYY-MM-DD, or a spreadsheet's YYYY-MM-DD HH:MM. */
function parseDate(value: string): string | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})(?:[ T].*)?$/.exec(value.trim());
  if (!match) return null;
  const [, y, m, d] = match;
  const date = new Date(Number(y), Number(m) - 1, Number(d));
  if (date.getMonth() !== Number(m) - 1 || date.getDate() !== Number(d)) return null;
  return `${y}-${m}-${d}`;
}

/** Subtasks in one cell, separated by semicolons, pipes or line breaks. */
function splitSubtasks(value: string): string[] {
  return value
    .split(/[;|\n]/)
    .map((s) => s.trim())
    .filter(Boolean);
}

export function buildImportPreview(text: string, ctx: ImportContext): ImportPreview {
  const [header = [], ...body] = parseCsv(text);
  const columns = header.map(columnFor);
  const index = (column: ImportColumn) => columns.indexOf(column);
  const hoursColumn = header.findIndex((h) => normalizeHeader(h) === 'hours');
  const people = new Map(
    ctx.people.filter((p) => p.email).map((p) => [p.email!.trim().toLowerCase(), p] as const)
  );
  const statuses = new Map(ctx.statuses.map((s) => [s.name.trim().toLowerCase(), s] as const));
  const truncated = body.length > MAX_IMPORT_ROWS;

  const rows = body.slice(0, MAX_IMPORT_ROWS).map((cells, i): ImportPreviewRow => {
    const cell = (column: ImportColumn) => {
      const at = index(column);
      return at >= 0 ? (cells[at] ?? '').trim() : '';
    };
    const issues: ImportIssue[] = [];
    const row: ImportPreviewRow = { line: i + 1, title: cell('title'), subtasks: [], issues };
    if (!row.title) issues.push({ kind: 'noTitle' });

    const description = cells[index('description')] ?? '';
    if (index('description') >= 0 && description.trim()) row.description = description.trim();

    const assignee = cell('assignee');
    if (assignee) {
      const person = people.get(assignee.toLowerCase());
      if (person) {
        row.assigneeId = person.id;
        row.assigneeLabel = person.name;
      } else issues.push({ kind: 'unknownAssignee', value: assignee });
    }

    const due = cell('due');
    if (due) {
      const parsed = parseDate(due);
      if (parsed) row.dueDate = parsed;
      else issues.push({ kind: 'badDate', value: due });
    }

    const priority = cell('priority').toLowerCase();
    if (priority) {
      if (PRIORITIES.has(priority)) row.priority = priority;
      else issues.push({ kind: 'badPriority', value: cell('priority') });
    }

    const effort = cell('effort');
    if (effort) {
      const minutes = parseEffort(effort, index('effort') === hoursColumn);
      if (minutes !== null) row.effortMinutes = minutes;
      else issues.push({ kind: 'badEffort', value: effort });
    }

    const section = cell('section');
    if (section) {
      const status = statuses.get(section.toLowerCase());
      if (status) {
        row.sectionId = status.id;
        row.sectionLabel = status.name;
      } else issues.push({ kind: 'unknownSection', value: section });
    }

    row.subtasks = splitSubtasks(cells[index('subtasks')] ?? '');
    return row;
  });

  return {
    columns: columns.filter((c): c is ImportColumn => c !== undefined),
    rows,
    missingTitle: index('title') < 0,
    truncated,
  };
}
