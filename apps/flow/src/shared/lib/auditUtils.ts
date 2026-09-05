import type { ActivityEntry } from '../stores/flowStore';
import { toLocalIso } from './dateUtils';

/** Returns the most recent activity entry across all tasks. */
export function latestActivityEntry(activity: ActivityEntry[]): ActivityEntry | undefined {
  if (activity.length === 0) return undefined;
  return [...activity].sort((a, b) => b.createdAt.localeCompare(a.createdAt))[0];
}

/** Returns all activity entries within the last N days. */
export function recentActivity(
  activity: ActivityEntry[],
  days: number,
  now?: Date
): ActivityEntry[] {
  const cutoff = new Date(now ?? new Date());
  cutoff.setDate(cutoff.getDate() - days);
  const cutoffStr = toLocalIso(cutoff);
  return activity.filter((e) => e.createdAt >= cutoffStr);
}

/** Returns the count of activity entries by type. */
export function activityCountByType(
  activity: ActivityEntry[]
): Partial<Record<ActivityEntry['type'], number>> {
  const counts: Partial<Record<ActivityEntry['type'], number>> = {};
  for (const entry of activity) {
    counts[entry.type] = (counts[entry.type] ?? 0) + 1;
  }
  return counts;
}

/** Returns all unique actors from an activity list. */
export function uniqueActors(activity: ActivityEntry[]): string[] {
  return [...new Set(activity.map((e) => e.actor))];
}

/** Returns activity entries for a specific actor. */
export function activityForActor(activity: ActivityEntry[], actor: string): ActivityEntry[] {
  return activity.filter((e) => e.actor === actor);
}

/** Returns the most active actor (by number of entries). */
export function mostActiveActor(activity: ActivityEntry[]): string | null {
  if (activity.length === 0) return null;
  const counts = new Map<string, number>();
  for (const entry of activity) {
    counts.set(entry.actor, (counts.get(entry.actor) ?? 0) + 1);
  }
  let maxCount = 0;
  let maxActor = '';
  counts.forEach((count, actor) => {
    if (count > maxCount) {
      maxCount = count;
      maxActor = actor;
    }
  });
  return maxActor;
}

/** Returns the date string of the first recorded activity. */
export function firstActivityDate(activity: ActivityEntry[]): string | null {
  if (activity.length === 0) return null;
  return [...activity].sort((a, b) => a.createdAt.localeCompare(b.createdAt))[0].createdAt;
}
