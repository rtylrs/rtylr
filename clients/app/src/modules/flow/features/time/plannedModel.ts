/**
 * Team hours > Planned (T-341): the people-by-days grid built from the planned
 * workload response, and how each cell compares with the daily target.
 */

export type WorkWeek = 'mon_fri' | 'sun_thu';

/** ISO weekdays (1 Monday … 7 Sunday) for the API, and JS getDay() values for the grid. */
export const WORK_WEEKS: Record<WorkWeek, { iso: string; jsDays: number[] }> = {
  mon_fri: { iso: '1,2,3,4,5', jsDays: [1, 2, 3, 4, 5] },
  sun_thu: { iso: '7,1,2,3,4', jsDays: [0, 1, 2, 3, 4] },
};

export type CellTone = 'off' | 'under' | 'ok' | 'over';

export interface PlannedCell {
  date: string;
  minutes: number;
  taskIds: string[];
  working: boolean;
  tone: CellTone;
}

export interface PlannedRow {
  userId: string;
  name: string;
  cells: PlannedCell[];
  totalMinutes: number;
}

export interface PlannedResponse {
  days?: Array<{ userId?: string; date?: string; plannedMinutes?: number; taskIds?: string[] }>;
  tasks?: Array<{ assigneeId?: string; assigneeName?: string }>;
  everyone?: boolean;
  people?: string[];
}

/**
 * Under the target is amber, more than an hour over it is red, and a day off
 * with nothing planned is blank. Within an hour over counts as on target, so a
 * 5h target accepts up to 6h.
 */
export function cellTone(minutes: number, targetMinutes: number, working: boolean): CellTone {
  if (!working && minutes === 0) return 'off';
  if (minutes < targetMinutes) return working ? 'under' : 'ok';
  if (minutes > targetMinutes + 60) return 'over';
  return 'ok';
}

const weekday = (date: string): number =>
  new Date(
    Number(date.slice(0, 4)),
    Number(date.slice(5, 7)) - 1,
    Number(date.slice(8, 10))
  ).getDay();

/**
 * One row per person the caller may see: everyone on the team when the
 * response says so, otherwise the people it lists. Anyone with planned work
 * but missing from the team list still gets a row.
 */
export function buildPlannedRows(
  response: PlannedResponse | undefined,
  team: ReadonlyArray<{ id: string; name: string }>,
  dates: readonly string[],
  workWeek: WorkWeek,
  targetMinutes: number
): PlannedRow[] {
  const byUserDay = new Map<string, { minutes: number; taskIds: string[] }>();
  for (const day of response?.days ?? []) {
    if (!day.userId || !day.date) continue;
    byUserDay.set(`${day.userId}|${day.date}`, {
      minutes: day.plannedMinutes ?? 0,
      taskIds: (day.taskIds ?? []).map(String),
    });
  }
  // The team list names everyone it holds; the response names anyone with
  // planned work, which covers people the caller cannot list.
  const names = new Map<string, string>();
  for (const task of response?.tasks ?? []) {
    if (task.assigneeId && task.assigneeName) names.set(String(task.assigneeId), task.assigneeName);
  }
  for (const m of team) names.set(String(m.id), m.name);
  const visible = response?.everyone
    ? team.map((m) => String(m.id))
    : (response?.people ?? []).map(String);
  const ids = [
    ...new Set([...visible, ...(response?.days ?? []).map((d) => String(d.userId ?? ''))]),
  ].filter(Boolean);
  const working = new Set(WORK_WEEKS[workWeek].jsDays);

  return ids
    .map((userId): PlannedRow => {
      const cells = dates.map((date): PlannedCell => {
        const entry = byUserDay.get(`${userId}|${date}`);
        const minutes = entry?.minutes ?? 0;
        const isWorking = working.has(weekday(date));
        return {
          date,
          minutes,
          taskIds: entry?.taskIds ?? [],
          working: isWorking,
          tone: cellTone(minutes, targetMinutes, isWorking),
        };
      });
      return {
        userId,
        name: names.get(userId) ?? userId,
        cells,
        totalMinutes: cells.reduce((sum, c) => sum + c.minutes, 0),
      };
    })
    .sort((a, b) => a.name.localeCompare(b.name));
}

/** Hours with at most one decimal: 5, 2.5, 0.8. */
export function formatHours(minutes: number): string {
  return String(Math.round((minutes / 60) * 10) / 10);
}
