import type { Task } from '../../shared/stores/flowStore';
import type { BoardGroupKey } from '../../shared/lib/boardUtils';
import { addDays, todayIso } from '../../shared/lib/dateUtils';

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

export function laneKeyFor(t: Task, groupBy: BoardGroupKey): string {
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
