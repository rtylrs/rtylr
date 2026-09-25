/**
 * The calendar day a Date falls on where the reader is standing, YYYY-MM-DD.
 *
 * Local, not UTC. `toISOString()` gives the UTC day, which at UTC+3 is the
 * previous day for the first three hours of every local day, and the whole day
 * for a Date built from local parts.
 */
export function toLocalIso(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/** Today as YYYY-MM-DD in the reader's own timezone. */
export function todayIso(): string {
  return toLocalIso(new Date());
}
