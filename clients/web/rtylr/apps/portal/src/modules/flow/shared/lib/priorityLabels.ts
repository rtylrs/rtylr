import type { Task } from '../stores/flowStore';

/**
 * The i18n key for a priority.
 *
 * `PRIORITY_LABELS` in the store is a hardcoded English map, and it was what
 * the priority glyph's tooltip and QuickCreate's `aria-label` read, so both
 * announced English in Arabic and French. `locales/*​/flowPriority.json` has had
 * the translations all along; only the mapping was missing.
 *
 * The one irregular name is `no_priority`, which is `noPriority` in the locale
 * files, so a plain interpolation of the priority value does not resolve.
 */
const KEYS: Record<Task['priority'], string> = {
  urgent: 'flowPriority.urgent',
  high: 'flowPriority.high',
  medium: 'flowPriority.medium',
  low: 'flowPriority.low',
  no_priority: 'flowPriority.noPriority',
};

export function priorityLabelKey(priority: Task['priority']): string {
  return KEYS[priority] ?? 'flowPriority.noPriority';
}
