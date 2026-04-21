import React, { useMemo } from 'react';
import { PieChart } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import api from '../../../../shared/lib/api';

/**
 * Revenue split by app surface. Hits the new `/insights/revenue/by-app`
 * endpoint that landed this session — one horizontal bar per app
 * (POS, Finance, ERP, CRM) sized by its share of MTD revenue.
 *
 * Keeping this deliberately thin: no date picker, no toggles, no
 * drill-down. That belongs in the Insights app. Here it's a hero-
 * tile summary.
 */

interface AppSlice {
  app: string;
  amountMinorUnits: number;
  pct: number;
}

interface ApiResponse {
  by_app?: {
    app: string;
    amount_minor_units: number;
    pct: number;
  }[];
}

const APP_COLOUR: Record<string, string> = {
  pos: 'var(--accent)',
  finance: '#a78bfa',
  erp: '#34d399',
  crm: '#f59e0b',
  other: 'var(--dash-text-dim)',
};

function formatMoney(minor: number): string {
  const major = minor / 100;
  if (major >= 1_000_000) return `$${(major / 1_000_000).toFixed(1)}M`;
  if (major >= 1_000) return `$${(major / 1_000).toFixed(1)}K`;
  return `$${major.toFixed(0)}`;
}

const RevenueByAppWidget: React.FC = (): React.JSX.Element => {
  const { data: raw } = useQuery({
    queryKey: ['insights-revenue-by-app'],
    queryFn: async (): Promise<AppSlice[]> => {
      try {
        const res = await api.get<ApiResponse>('/insights/revenue/by-app');
        return (res.data.by_app ?? []).map((s) => ({
          app: s.app,
          amountMinorUnits: s.amount_minor_units,
          pct: s.pct,
        }));
      } catch {
        // Insights namespace may not be deployed everywhere yet —
        // fail open to the empty-state render below.
        return [];
      }
    },
    staleTime: 60_000,
  });

  const slices = useMemo<AppSlice[]>(() => raw ?? [], [raw]);
  const total = slices.reduce((acc, s) => acc + s.amountMinorUnits, 0);

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <PieChart size={14} style={{ color: 'var(--accent)' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Revenue by app (MTD)
        </p>
      </div>
      {slices.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center text-center p-4">
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            Start selling to see the split across your apps.
          </p>
        </div>
      ) : (
        <div className="flex-1 flex flex-col gap-2 p-4">
          <div className="flex items-baseline justify-between">
            <span
              className="text-[10px] uppercase tracking-wider"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              Total
            </span>
            <span className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
              {formatMoney(total)}
            </span>
          </div>
          <div className="flex flex-col gap-2 mt-2">
            {slices.map((s) => {
              const colour = APP_COLOUR[s.app.toLowerCase()] ?? APP_COLOUR.other;
              const width = Math.max(4, Math.min(100, s.pct));
              return (
                <div key={s.app} className="flex flex-col gap-1">
                  <div className="flex items-center justify-between text-[11px]">
                    <span
                      style={{ color: 'var(--dash-text)' }}
                      className="font-medium uppercase tracking-wider"
                    >
                      {s.app}
                    </span>
                    <span style={{ color: 'var(--dash-text-dim)' }}>
                      {formatMoney(s.amountMinorUnits)} · {s.pct.toFixed(0)}%
                    </span>
                  </div>
                  <div
                    className="w-full h-1.5 rounded-full overflow-hidden"
                    style={{ background: 'var(--dash-border)' }}
                  >
                    <div
                      className="h-full rounded-full transition-all"
                      style={{ width: `${width}%`, background: colour }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

export default RevenueByAppWidget;
