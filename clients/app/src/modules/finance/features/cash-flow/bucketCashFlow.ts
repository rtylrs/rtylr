import type { CashFlowPoint } from '../../shared/components/CashFlowChart';

export type CashFlowPeriod = 'month' | 'quarter' | 'year';

export interface CashFlowBucket extends CashFlowPoint {
  /** Short label under the bar, empty to leave a gap in a dense axis. */
  label: string;
  /** Full name of the bucket for the tooltip. */
  title: string;
}

const parse = (iso: string): Date => {
  const [y, m, d] = iso.split('-').map(Number);
  return new Date(y, m - 1, d);
};
const fmt = (date: Date, opts: Intl.DateTimeFormatOptions): string =>
  date.toLocaleDateString('en-US', opts);

/**
 * The analytics API answers one point per day, so a year was 365 points
 * squeezed into one chart with unreadable labels. A month stays daily, a
 * quarter groups by week (Monday start) and a year by month.
 */
export function bucketCashFlow(points: CashFlowPoint[], period: CashFlowPeriod): CashFlowBucket[] {
  if (period === 'month') {
    return points.map((p) => {
      const date = parse(p.month);
      const day = date.getDate();
      return {
        ...p,
        label: day === 1 || day % 5 === 0 ? String(day) : '',
        title: fmt(date, { month: 'short', day: 'numeric' }),
      };
    });
  }
  const buckets = new Map<string, CashFlowBucket>();
  for (const p of points) {
    const date = parse(p.month);
    let key: string;
    let label: string;
    let title: string;
    if (period === 'year') {
      key = `${date.getFullYear()}-${date.getMonth()}`;
      label = fmt(date, { month: 'short' });
      title = fmt(date, { month: 'long', year: 'numeric' });
    } else {
      const monday = new Date(date);
      monday.setDate(date.getDate() - ((date.getDay() + 6) % 7));
      key = monday.toDateString();
      label = fmt(monday, { month: 'short', day: 'numeric' });
      title = `Week of ${label}`;
    }
    const bucket = buckets.get(key) ?? { month: key, inflow: 0, outflow: 0, label, title };
    bucket.inflow += p.inflow;
    bucket.outflow += p.outflow;
    buckets.set(key, bucket);
  }
  return [...buckets.values()];
}
