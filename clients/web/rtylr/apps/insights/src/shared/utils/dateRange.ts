// Shared date-range defaults for /insights/* endpoints that accept {from, to}
// query params. Backend defaults to today−30d → today when either param is
// unset; the frontend mirrors this so empty inputs round-trip to the same
// series the backend would pick.

export interface DateRangeParams {
  from?: string;
  to?: string;
}

// Local calendar day, not the UTC day. The reporting queries bucket by the
// dates they are handed, so a UTC boundary labelled "Today" showed yesterday's
// figures for the first hours of the local day (rtylr#1201).
function toISODate(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export function resolveDateRange(params: DateRangeParams | undefined): {
  from: string;
  to: string;
} {
  const now = new Date();
  const to = params?.to ?? toISODate(now);
  const fromDate = params?.from
    ? new Date(params.from)
    : new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
  const from = params?.from ?? toISODate(fromDate);
  return { from, to };
}
