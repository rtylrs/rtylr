// Pure helpers for manager override validation and display.

import type { ManagerOverride } from '../types/pos';

/** Return true if a manager override was created within the last N minutes. */
export function isOverrideRecent(override: ManagerOverride, withinMinutes = 10): boolean {
  const created = new Date(override.createdAt).getTime();
  if (!Number.isFinite(created)) return false;
  const now = Date.now();
  const diffMs = now - created;
  return diffMs >= 0 && diffMs <= withinMinutes * 60 * 1000;
}

/** Return a human-readable summary of the override. */
export function overrideSummary(override: ManagerOverride): string {
  return `Manager ${override.managerId} authorized "${override.action}" (${override.reason})`;
}

/** Return true if the reason string is non-empty and meaningful. */
export function isValidOverrideReason(reason: string): boolean {
  return reason.trim().length >= 3;
}

/** Group overrides by the action type. */
export function overridesByAction(
  overrides: ManagerOverride[],
): Record<string, ManagerOverride[]> {
  const result: Record<string, ManagerOverride[]> = {};
  for (const o of overrides) {
    if (!result[o.action]) result[o.action] = [];
    result[o.action].push(o);
  }
  return result;
}

/** Sort overrides by createdAt descending (newest first). */
export function sortOverridesNewestFirst(overrides: ManagerOverride[]): ManagerOverride[] {
  return [...overrides].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  );
}

/** Return the count of overrides per manager. */
export function overrideCountByManager(overrides: ManagerOverride[]): Record<string, number> {
  const result: Record<string, number> = {};
  for (const o of overrides) {
    result[o.managerId] = (result[o.managerId] ?? 0) + 1;
  }
  return result;
}
