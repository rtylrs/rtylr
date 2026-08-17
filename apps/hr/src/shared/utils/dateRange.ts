/** Local-date (not UTC) formatting so week boundaries match the user's own calendar day. */
export function toLocalDateString(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export interface WeekRange {
  start: string;
  end: string;
}

/** Current Monday-Sunday week, as local YYYY-MM-DD strings. */
export function getWeekRange(): WeekRange {
  const now = new Date();
  const day = now.getDay();
  const diff = day === 0 ? 6 : day - 1; // Monday-based
  const monday = new Date(now.getFullYear(), now.getMonth(), now.getDate() - diff);
  const sunday = new Date(now.getFullYear(), now.getMonth(), now.getDate() - diff + 6);
  return {
    start: toLocalDateString(monday),
    end: toLocalDateString(sunday),
  };
}
