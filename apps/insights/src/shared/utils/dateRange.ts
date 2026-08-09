// Shared date-range defaults for /insights/* endpoints that accept {from, to}
// query params. Backend defaults to today−30d → today when either param is
// unset; the frontend mirrors this so empty inputs round-trip to the same
// series the backend would pick.

export interface DateRangeParams {
  from?: string;
  to?: string;
}

function toISODate(d: Date): string {
  const y = d.getUTCFullYear();
  const m = String(d.getUTCMonth() + 1).padStart(2, '0');
  const day = String(d.getUTCDate()).padStart(2, '0');
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
