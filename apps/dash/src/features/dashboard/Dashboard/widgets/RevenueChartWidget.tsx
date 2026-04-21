import React, { useMemo } from 'react';
import { BarChart3 } from 'lucide-react';
import { useRevenueSeries } from '../../../../shared/hooks/queries/useRevenueSeries';

/**
 * 7-day revenue bars. Data comes from `/insights/revenue` via
 * useRevenueSeries. Until the insights namespace is deployed to an
 * environment, the hook returns an empty array and the widget
 * renders the same no-data state it has today.
 */

const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

interface DayBucket {
  label: string;
  value: number;
}

const RevenueChartWidget: React.FC = (): React.JSX.Element => {
  const { data } = useRevenueSeries(7);

  // Normalise the series into seven buckets keyed by weekday. We
  // reshape here rather than in the hook so the hook stays a thin
  // fetch + the widget owns display concerns.
  const buckets = useMemo<DayBucket[]>(() => {
    const today = new Date();
    const out: DayBucket[] = [];
    for (let i = 6; i >= 0; i--) {
      const d = new Date(today);
      d.setDate(d.getDate() - i);
      const iso = d.toISOString().slice(0, 10);
      const match = data.find((p) => p.date === iso);
      // minor units → major for display. Division by 100 keeps the
      // widget bar-proportional; exact $ values surface in the
      // tooltip of the full analytics page, not here.
      const value = match ? match.amountMinorUnits / 100 : 0;
      out.push({ label: DAY_LABELS[d.getDay()], value });
    }
    return out;
  }, [data]);

  const max = Math.max(...buckets.map((b) => b.value), 1);
  const allZero = buckets.every((b) => b.value === 0);

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <BarChart3 size={14} style={{ color: 'var(--accent)' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Revenue (7 days)
        </p>
      </div>
      <div className="flex-1 flex items-end gap-2 px-4 pb-4 pt-2">
        {buckets.map((b, i) => (
          <div key={`${b.label}-${i}`} className="flex-1 flex flex-col items-center gap-1">
            <div className="w-full relative" style={{ height: '100%', minHeight: 60 }}>
              <div
                className="absolute bottom-0 w-full rounded-t-md transition-all"
                style={{
                  height: b.value > 0 ? `${(b.value / max) * 100}%` : '4px',
                  background: b.value > 0 ? 'var(--accent)' : 'var(--dash-border)',
                  minHeight: 4,
                }}
              />
            </div>
            <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
              {b.label}
            </span>
          </div>
        ))}
      </div>
      {allZero && (
        <p className="text-center text-[11px] pb-3" style={{ color: 'var(--dash-text-dim)' }}>
          Revenue will appear once orders start flowing
        </p>
      )}
    </div>
  );
};

export default RevenueChartWidget;
