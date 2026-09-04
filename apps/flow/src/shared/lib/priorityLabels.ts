import type { Task } from '../stores/flowStore';

/**
 * The i18n key for a priority.
 *
 * `PRIORITY_LABELS` in the store is a hardcoded English map, and it was what
 * the priority glyph's tooltip and QuickCreate's `aria-label` read, so both
 * announced English in Arabic and French. `locales/*​/priority.json` has had
 * the translations all along; only the mapping was missing.
 *
 * The one irregular name is `no_priority`, which is `noPriority` in the locale
 * files, so a plain interpolation of the priority value does not resolve.
 */
const KEYS: Record<Task['priority'], string> = {
  urgent: 'priority.urgent',
  high: 'priority.high',
  medium: 'priority.medium',
  low: 'priority.low',
  no_priority: 'priority.noPriority',
};

export function priorityLabelKey(priority: Task['priority']): string {
  return KEYS[priority] ?? 'priority.noPriority';
}
