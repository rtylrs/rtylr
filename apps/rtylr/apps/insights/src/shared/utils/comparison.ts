// Period comparison helpers for the insights revenue and financial views.
// These derive the "compare to previous period" date ranges that the
// backend uses when the user picks a comparison window.

export interface DateRange {
  from: string;
  to: string;
}

/**
 * previousPeriod shifts a date range backward by its own length so
 * the two windows are exactly equal in duration.
 * e.g. [Jan 1, Jan 31] (30 days) -> [Dec 2, Dec 31]
 */
export function previousPeriod(range: DateRange): DateRange {
  const from = new Date(range.from + 'T00:00:00Z');
  const to = new Date(range.to + 'T00:00:00Z');
  const lengthMs = to.getTime() - from.getTime() + 24 * 60 * 60 * 1000; // inclusive

  const prevTo = new Date(from.getTime() - 24 * 60 * 60 * 1000);
  const prevFrom = new Date(prevTo.getTime() - lengthMs + 24 * 60 * 60 * 1000);

  return {
    from: prevFrom.toISOString().slice(0, 10),
    to: prevTo.toISOString().slice(0, 10),
  };
}

/**
 * rangeLength returns the number of days in a date range, inclusive.
 */
export function rangeLength(range: DateRange): number {
  const from = new Date(range.from + 'T00:00:00Z').getTime();
  const to = new Date(range.to + 'T00:00:00Z').getTime();
  return Math.round((to - from) / (1000 * 60 * 60 * 24)) + 1;
}

/**
 * compareRanges returns the absolute and relative difference between
 * a current and a reference value.
 */
export function compareRanges(
  current: number,
  previous: number
): { absolute: number; pct: number } {
  const absolute = current - previous;
  const pct = previous === 0 ? 0 : ((current - previous) / Math.abs(previous)) * 100;
  return { absolute, pct };
}

/**
 * sameMonthLastYear returns the DateRange that corresponds to the same
 * calendar month in the previous year.
 */
export function sameMonthLastYear(range: DateRange): DateRange {
  const shiftYear = (iso: string) => (parseInt(iso.slice(0, 4)) - 1).toString() + iso.slice(4);
  return { from: shiftYear(range.from), to: shiftYear(range.to) };
}
