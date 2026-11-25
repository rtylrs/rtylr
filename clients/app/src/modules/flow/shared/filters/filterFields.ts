import type { TFunction } from 'i18next';
import type { Task } from '../stores/flowStore';
import { initialsFromName } from '../lib/memberUtils';
import { statusLabelKey } from '../lib/statusLabels';
import type { ApiField } from '../services/flowApi';
import { fieldKey, isListFilterable } from '../lib/customFields';
import {
  NO_FIELD_VALUE,
  PRIORITIES,
  UNASSIGNED,
  type CustomFilterField,
  type ListField,
} from './taskFilters';
import { NO_EPIC, type EpicSummary } from '../lib/taskHierarchy';

const BUILT_IN_STATUSES: Task['status'][] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
  'cancelled',
];

export interface FilterOption {
  value: string;
  /** Already translated, or a person's / label's own name. */
  label: string;
  /** A dot of this colour before the label (labels, statuses). */
  color?: string;
  /** Initials in a small avatar before the label (people). */
  initials?: string;
}

export interface FilterFieldConfig {
  field: ListField | CustomFilterField;
  label: string;
  options: FilterOption[];
}

export interface FilterPerson {
  id: string;
  name: string;
}

const PRIORITY_KEY: Record<Task['priority'], string> = {
  urgent: 'flowPriority.urgent',
  high: 'flowPriority.high',
  medium: 'flowPriority.medium',
  low: 'flowPriority.low',
  no_priority: 'flowPriority.noPriority',
};

export function priorityField(t: TFunction): FilterFieldConfig {
  return {
    field: 'priorities',
    label: t('flow.filters.fields.priority'),
    options: PRIORITIES.map((p) => ({ value: p, label: t(PRIORITY_KEY[p]) })),
  };
}

/** The built-in statuses, for lists that span projects with their own columns. */
export function builtInStatusField(t: TFunction): FilterFieldConfig {
  return {
    field: 'statuses',
    label: t('flow.filters.fields.status'),
    options: BUILT_IN_STATUSES.map((s) => ({ value: s, label: t(statusLabelKey(s)) })),
  };
}

/** People, the current user first and marked, then Unassigned. */
export function assigneeField(
  people: readonly FilterPerson[],
  currentUserId: string | null,
  t: TFunction
): FilterFieldConfig {
  return {
    field: 'assignees',
    label: t('flow.filters.fields.assignee'),
    options: [
      ...orderPeople(people, currentUserId).map((p) => ({
        value: p.id,
        label: p.id === currentUserId ? t('flow.filters.you', { name: p.name }) : p.name,
        initials: initialsFromName(p.name),
      })),
      { value: UNASSIGNED, label: t('flow.filters.unassigned') },
    ],
  };
}

export function labelField(
  labels: readonly { id: number | string; name: string; color: string }[],
  t: TFunction
): FilterFieldConfig {
  return {
    field: 'labels',
    label: t('flow.filters.fields.label'),
    options: labels.map((l) => ({ value: String(l.id), label: l.name, color: l.color })),
  };
}

/** A project's epics in board order, then "No epic". */
export function epicField(epics: readonly EpicSummary[], t: TFunction): FilterFieldConfig {
  return {
    field: 'epics',
    label: t('flow.filters.fields.epic'),
    options: [
      ...epics.map((e) => ({ value: e.id, label: e.title, color: e.color })),
      { value: NO_EPIC, label: t('flow.epics.noEpic') },
    ],
  };
}

/** The CI state of a task's open pull requests (T-376). */
export function checksField(t: TFunction): FilterFieldConfig {
  return {
    field: 'checks',
    label: t('flow.filters.fields.checks'),
    options: (['failing', 'running', 'passing'] as const).map((value) => ({
      value,
      label: t(`flow.checks.state.${value}`),
    })),
  };
}

export function namedField(
  field: Extract<ListField, 'statuses' | 'cycles' | 'milestones' | 'projects'>,
  label: string,
  items: readonly { id: string; name: string; color?: string }[]
): FilterFieldConfig {
  return {
    field,
    label,
    options: items.map((i) => ({ value: String(i.id), label: i.name, color: i.color })),
  };
}

/** Me first, then everyone else alphabetically, so the strip is stable. */
export function orderPeople(
  people: readonly FilterPerson[],
  currentUserId: string | null
): FilterPerson[] {
  const seen = new Set<string>();
  const unique = people.filter((p) => {
    if (!p.id || seen.has(p.id)) return false;
    seen.add(p.id);
    return true;
  });
  return unique.sort((a, b) => {
    if (a.id === currentUserId) return -1;
    if (b.id === currentUserId) return 1;
    return a.name.localeCompare(b.name);
  });
}

/**
 * Members plus anyone a task is still assigned to who is no longer a member,
 * so a removed teammate's tasks can still be filtered.
 */
export function peopleFrom(
  members: readonly { userId: string; name: string }[],
  tasks: readonly Pick<Task, 'assigneeId' | 'assignee'>[]
): FilterPerson[] {
  const byId = new Map<string, string>();
  for (const m of members) byId.set(String(m.userId), m.name);
  for (const task of tasks) {
    if (task.assigneeId && !byId.has(task.assigneeId)) {
      byId.set(task.assigneeId, task.assignee || task.assigneeId);
    }
  }
  return [...byId].map(([id, name]) => ({ id, name }));
}

/**
 * A project's custom fields that filter as a checklist: selects, people and
 * checkboxes. Each lists its values plus "No value", so a task missing the
 * field can be found as easily as one holding it.
 */
export function customFieldFilters(
  fields: readonly ApiField[],
  people: readonly FilterPerson[],
  t: TFunction
): FilterFieldConfig[] {
  const none = { value: NO_FIELD_VALUE, label: t('flow.fields.noValue') };
  return fields
    .filter((f) => isListFilterable(f.type))
    .map((f) => {
      let options: FilterOption[];
      if (f.type === 'checkbox') {
        options = [
          { value: 'true', label: t('flow.fields.checked') },
          { value: NO_FIELD_VALUE, label: t('flow.fields.unchecked') },
        ];
      } else if (f.type === 'person') {
        options = [
          ...people.map((p) => ({
            value: p.id,
            label: p.name,
            initials: initialsFromName(p.name),
          })),
          none,
        ];
      } else {
        options = [
          ...f.options.map((o) => ({ value: o.id, label: o.label, color: o.color })),
          none,
        ];
      }
      return { field: fieldKey(f.id), label: f.name, options };
    });
}
