import type { Task } from '../../shared/stores/flowStore';
import type { BoardGroupKey } from '../../shared/lib/boardUtils';
import { addDays, todayIso } from '../../shared/lib/dateUtils';
import type { ApiField } from '../../shared/services/flowApi';
import { fieldIdFromKey, fieldKey, fieldValueParts } from '../../shared/lib/customFields';

export const PRIORITY_ORDER: Task['priority'][] = [
  'urgent',
  'high',
  'medium',
  'low',
  'no_priority',
];

export const PRIORITY_LABEL_KEY: Record<Task['priority'], string> = {
  urgent: 'flowPriority.urgent',
  high: 'flowPriority.high',
  medium: 'flowPriority.medium',
  low: 'flowPriority.low',
  no_priority: 'flowPriority.noPriority',
};

// The keys are lane identities used for bucketing, not text; translating them
// would break the grouping.
export const DUE_LANES: { key: string; labelKey: string }[] = [
  { key: 'Overdue', labelKey: 'flow.board.dueLane.overdue' },
  { key: 'Today', labelKey: 'flow.board.dueLane.today' },
  { key: 'This week', labelKey: 'flow.board.dueLane.thisWeek' },
  { key: 'This month', labelKey: 'flow.board.dueLane.thisMonth' },
  { key: 'Later', labelKey: 'flow.board.dueLane.later' },
  { key: 'No due date', labelKey: 'flow.board.dueLane.noDueDate' },
];

/** The lane of tasks with no value in the field the board is grouped by. */
export const NO_FIELD_LANE = '_no_value';

/** Custom fields a board can be grouped by: single selects, one lane per option. */
export function groupableFields(fields: readonly ApiField[]): ApiField[] {
  return fields.filter((f) => f.type === 'select');
}

/** A grouped-by field's lanes in option order, then "No value". */
export function fieldLanes(field: ApiField): { key: string; label?: string; labelKey?: string }[] {
  return [
    ...field.options.map((o) => ({ key: o.id, label: o.label })),
    { key: NO_FIELD_LANE, labelKey: 'flow.fields.noValue' },
  ];
}

/** The group-by choices a project's fields add after the built-in ones. */
export function fieldGroupOptions(
  fields: readonly ApiField[]
): { key: BoardGroupKey; label: string }[] {
  return groupableFields(fields).map((f) => ({ key: fieldKey(f.id), label: f.name }));
}

export function laneKeyFor(t: Task, groupBy: BoardGroupKey): string {
  const fieldId = fieldIdFromKey(groupBy);
  if (fieldId !== null) return fieldValueParts(t.fieldValues?.[fieldId])[0] ?? NO_FIELD_LANE;
  if (groupBy === 'assignee') return t.assignee || 'Unassigned';
  if (groupBy === 'priority') return t.priority;
  if (groupBy === 'due') {
    if (!t.dueDate) return 'No due date';
    // A due date is a calendar day, not an instant. Parsing the bare date as
    // UTC midnight and rounding against a local `now` slid a task due today
    // into "Overdue" late in the day (rtylr#1077), so compare day strings.
    const due = t.dueDate.split('T')[0] ?? '';
    if (due === '') return 'No due date';
    const today = todayIso();
    if (due < today) return 'Overdue';
    if (due === today) return 'Today';
    if (due <= addDays(today, 7)) return 'This week';
    if (due <= addDays(today, 30)) return 'This month';
    return 'Later';
  }
  return '_all';
}
