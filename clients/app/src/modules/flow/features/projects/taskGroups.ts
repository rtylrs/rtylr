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

/** The lane for cards with no label, cycle or milestone. */
export const NO_VALUE_LANE = '_none';

/** Groupings whose lanes are this project's labels, cycles or milestones. */
export type ValueGroupKey = Extract<BoardGroupKey, 'label' | 'cycle' | 'milestone'>;

export const NO_VALUE_LABEL_KEY: Record<ValueGroupKey, string> = {
  label: 'flow.board.noLabel',
  cycle: 'flow.board.noCycle',
  milestone: 'flow.board.noMilestone',
};

export const isValueGroupKey = (groupBy: BoardGroupKey): groupBy is ValueGroupKey =>
  groupBy === 'label' || groupBy === 'cycle' || groupBy === 'milestone';

/**
 * Every lane a card belongs in. One, except for labels: a card sits in the
 * lane of each label it carries. An id the board does not know (a label,
 * cycle or milestone since deleted) counts as no value, as it does on the card.
 */
export function laneKeysFor(
  t: Task,
  groupBy: BoardGroupKey,
  known?: ReadonlySet<string>
): string[] {
  if (!isValueGroupKey(groupBy)) return [laneKeyFor(t, groupBy)];
  const raw =
    groupBy === 'label'
      ? (t.labelIds ?? []).map(String)
      : [(groupBy === 'cycle' ? t.cycleId : t.milestoneId) ?? ''];
  const ids = [...new Set(raw.filter((id) => id !== '' && (!known || known.has(id))))];
  return ids.length > 0 ? ids : [NO_VALUE_LANE];
}

/**
 * The lanes for a label, cycle or milestone grouping: the options some card
 * uses, in the options' own order, then "No ..." when a card has none. Lanes
 * nobody uses are left out, as the assignee grouping leaves out people.
 */
export function valueLanes(
  tasks: readonly Task[],
  groupBy: ValueGroupKey,
  options: readonly { id: string | number; name: string }[]
): { key: string; label?: string; labelKey?: string }[] {
  const known = new Set(options.map((o) => String(o.id)));
  const used = new Set(tasks.flatMap((t) => laneKeysFor(t, groupBy, known)));
  const lanes: { key: string; label?: string; labelKey?: string }[] = options
    .filter((o) => used.has(String(o.id)))
    .map((o) => ({ key: String(o.id), label: o.name }));
  if (used.has(NO_VALUE_LANE)) {
    lanes.push({ key: NO_VALUE_LANE, labelKey: NO_VALUE_LABEL_KEY[groupBy] });
  }
  return lanes;
}
