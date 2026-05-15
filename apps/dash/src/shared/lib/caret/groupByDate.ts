import type { CaretChatSummary } from '@voxire/caret';

const MS_DAY = 86_400_000;

export type DateGroup = 'today' | 'yesterday' | 'last7' | 'last30' | 'older';

export interface GroupedChats {
  group: DateGroup;
  label: string;
  chats: CaretChatSummary[];
}

const LABELS: Record<DateGroup, string> = {
  today: 'Today',
  yesterday: 'Yesterday',
  last7: 'Previous 7 days',
  last30: 'Previous 30 days',
  older: 'Older',
};

function classify(updatedAt: string, now: number): DateGroup {
  const t = new Date(updatedAt).getTime();
  if (Number.isNaN(t)) return 'older';
  const startOfToday = new Date(now);
  startOfToday.setHours(0, 0, 0, 0);
  const startMs = startOfToday.getTime();
  if (t >= startMs) return 'today';
  if (t >= startMs - MS_DAY) return 'yesterday';
  if (t >= startMs - 7 * MS_DAY) return 'last7';
  if (t >= startMs - 30 * MS_DAY) return 'last30';
  return 'older';
}

/** Group chats into Today / Yesterday / 7d / 30d / Older buckets. */
export function groupChatsByDate(
  chats: CaretChatSummary[],
  now: number = Date.now()
): GroupedChats[] {
  const buckets: Record<DateGroup, CaretChatSummary[]> = {
    today: [],
    yesterday: [],
    last7: [],
    last30: [],
    older: [],
  };
  for (const c of chats) {
    buckets[classify(c.updatedAt, now)].push(c);
  }
  return (Object.keys(buckets) as DateGroup[])
    .map((g) => ({ group: g, label: LABELS[g], chats: buckets[g] }))
    .filter((g) => g.chats.length > 0);
}
