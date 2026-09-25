import React from 'react';
import { AlertTriangle, ChartNoAxesCombined } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import Skeleton from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { useDaySummary } from './useDaySummary';
import { formatCurrency } from '../../shared/utils/format';

const Stat: React.FC<{ label: string; value: string; tone?: 'positive' | 'negative' }> = ({
  label,
  value,
  tone,
}) => (
  <div className="p-4 rounded-xl bg-pos-card border border-pos-border">
    <div className="text-xs text-pos-muted uppercase tracking-wider">{label}</div>
    <div
      className={`mt-1 text-2xl font-bold ${
        tone === 'negative'
          ? 'text-destructive'
          : tone === 'positive'
            ? 'text-emerald-300'
            : 'text-pos-text'
      }`}
    >
      {value}
    </div>
  </div>
);

const DaySummaryScreen: React.FC = () => {
  const { summary, loading, error } = useDaySummary();

  return (
    <ScreenLayout title="Day Summary" subtitle={summary?.date}>
      <div className="p-5">
        {loading ? (
          <Skeleton rows={8} />
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Unable to load" message={error} />
        ) : summary ? (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3 max-w-5xl">
            <Stat label="Gross sales" value={formatCurrency(summary.grossSales)} tone="positive" />
            <Stat label="Net sales" value={formatCurrency(summary.netSales)} />
            <Stat label="Tax" value={formatCurrency(summary.taxCollected)} />
            <Stat label="Tips" value={formatCurrency(summary.tips)} tone="positive" />
            <Stat label="Refunds" value={formatCurrency(summary.refunds)} tone="negative" />
            <Stat label="Voids" value={formatCurrency(summary.voids)} tone="negative" />
            <Stat label="Orders" value={String(summary.orderCount)} />
            <Stat label="Avg ticket" value={formatCurrency(summary.averageTicket)} />
            <Stat label="Opening float" value={formatCurrency(summary.openingFloat)} />
            <Stat label="Cash drops" value={formatCurrency(summary.cashDrops)} />
            <Stat label="Expected cash" value={formatCurrency(summary.closingCash)} />
          </div>
        ) : (
          <EmptyState icon={ChartNoAxesCombined} title="No data" />
        )}
      </div>
    </ScreenLayout>
  );
};

export default DaySummaryScreen;
