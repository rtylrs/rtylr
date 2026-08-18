import React from 'react';
import { Clock } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { toFiniteNumber, unwrap } from '@voxire/api';
import { safe } from '../../../../shared/lib/api';
import WidgetHeader from './WidgetHeader';

/**
 * Invoice aging buckets - accounts-receivable at a glance. Hits
 * `/insights/financial/invoice-aging`, renders four stacked buckets:
 * 0-30d, 31-60d, 61-90d, 90+ overdue. Dollar totals + invoice counts.
 *
 * Most businesses will stare at the 90+ bucket - that's the column
 * styled red. The healthy 0-30 sits in muted green.
 */

interface Bucket {
  label: string;
  amountMinorUnits: number;
  count: number;
}

interface ApiResponse {
  buckets?: {
    label: string;
    amountMinorUnits?: number | string;
    amount_minor_units?: number | string;
    count: number | string;
  }[];
}

const BUCKET_COLOUR: Record<string, string> = {
  '0-30': 'var(--aging-current)',
  '31-60': 'var(--aging-warning)',
  '61-90': 'var(--aging-overdue)',
  '90+': 'var(--aging-critical)',
};

function formatMoney(minor: number): string {
  const major = minor / 100;
  if (major >= 1_000_000) return `$${(major / 1_000_000).toFixed(1)}M`;
  if (major >= 1_000) return `$${(major / 1_000).toFixed(1)}K`;
  return `$${major.toFixed(0)}`;
}

const InvoiceAgingWidget: React.FC = (): React.JSX.Element => {
  const { data: raw, isError } = useQuery({
    queryKey: ['insights-invoice-aging'],
    queryFn: async (): Promise<Bucket[]> => {
      const res = await unwrap(safe.get<ApiResponse>('/insights/financial/v1/invoice-aging'));
      return (res.buckets ?? []).map((b) => ({
        label: b.label,
        amountMinorUnits: toFiniteNumber(b.amountMinorUnits ?? b.amount_minor_units),
        count: toFiniteNumber(b.count),
      }));
    },
    staleTime: 60_000,
  });

  const buckets = raw ?? [];
  const allEmpty = buckets.every((b) => b.count === 0);

  return (
    <div className="flex flex-col h-full">
      <WidgetHeader icon={Clock} title="Invoice aging" />
      {isError ? (
        <div className="flex-1 flex flex-col items-center justify-center text-center p-4">
          <p className="text-xs text-red-500">Failed to load invoice aging data.</p>
        </div>
      ) : buckets.length === 0 || allEmpty ? (
        <div className="flex-1 flex flex-col items-center justify-center text-center p-4">
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            All outstanding invoices current.
          </p>
        </div>
      ) : (
        <div className="flex-1 flex flex-col gap-2 p-4">
          {buckets.map((b) => {
            const colour = BUCKET_COLOUR[b.label] ?? 'var(--dash-text-dim)';
            return (
              <div
                key={b.label}
                className="flex items-center justify-between px-3 py-2 rounded-lg"
                style={{ background: 'var(--dash-bg)' }}
              >
                <div className="flex items-center gap-2">
                  <span
                    className="w-2 h-2 rounded-full"
                    style={{ background: colour }}
                    aria-hidden="true"
                  />
                  <span
                    className="text-[11px] font-medium uppercase tracking-wider"
                    style={{ color: 'var(--dash-text)' }}
                  >
                    {b.label} days
                  </span>
                </div>
                <div className="flex items-baseline gap-2">
                  <span className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {formatMoney(b.amountMinorUnits)}
                  </span>
                  <span className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
                    {b.count} invoice{b.count === 1 ? '' : 's'}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default InvoiceAgingWidget;
