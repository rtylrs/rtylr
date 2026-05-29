/**
 * Activity feed utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type ActivityType =
  | 'create'
  | 'update'
  | 'delete'
  | 'login'
  | 'logout'
  | 'export'
  | 'import'
  | 'comment'
  | 'mention'
  | 'assign';

export interface ActivityEntry {
  id: string;
  type: ActivityType;
  actorName: string;
  targetName?: string;
  createdAt: string;
  read: boolean;
}

const ACTIVITY_LABELS: Record<ActivityType, string> = {
  create: 'Created',
  update: 'Updated',
  delete: 'Deleted',
  login: 'Logged in',
  logout: 'Logged out',
  export: 'Exported',
  import: 'Imported',
  comment: 'Commented',
  mention: 'Mentioned',
  assign: 'Assigned',
};

/** Return a human-readable label for an activity type. */
export function activityLabel(type: ActivityType): string {
  return ACTIVITY_LABELS[type] ?? type;
}

/** Build a short description for an activity entry. */
export function activityDescription(entry: ActivityEntry): string {
  const action = activityLabel(entry.type);
  if (entry.targetName) return `${entry.actorName} ${action.toLowerCase()} ${entry.targetName}`;
  return `${entry.actorName} ${action.toLowerCase()}`;
}

/** Return only unread activity entries. */
export function unreadActivities(entries: ActivityEntry[]): ActivityEntry[] {
  return entries.filter((e) => !e.read);
}

/** Return the count of unread activity entries. */
export function unreadActivityCount(entries: ActivityEntry[]): number {
  return unreadActivities(entries).length;
}

/** Filter entries by activity type. */
export function filterByType(entries: ActivityEntry[], type: ActivityType): ActivityEntry[] {
  return entries.filter((e) => e.type === type);
}

/** Sort entries by createdAt descending (newest first). Does not mutate. */
export function sortByDate(entries: ActivityEntry[]): ActivityEntry[] {
  return [...entries].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/** Return entries from the last N days. */
export function recentEntries(
  entries: ActivityEntry[],
  days: number,
  now = new Date()
): ActivityEntry[] {
  const cutoff = now.getTime() - days * 86_400_000;
  return entries.filter((e) => new Date(e.createdAt).getTime() >= cutoff);
}
