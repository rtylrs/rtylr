import React from 'react';
import { DollarSign, ShoppingBag, Users, TrendingUp, FileText } from 'lucide-react';
import { Skeleton } from '../../../../shared/components/ui';
import { useKpis } from '../../../../shared/hooks/queries/useKpis';
import { useOpenInvoiceCount } from '../../../../shared/hooks/queries/useOpenInvoiceCount';

function fmt(n: number | undefined | null) {
  if (n == null) return '—';
  return `$${n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
}

function KpiCard({
  label,
  value,
  sub,
  icon: Icon,
  color,
  isLoading,
}: {
  label: string;
  value: React.ReactNode;
  sub: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  color: string;
  isLoading?: boolean;
}) {
  return (
    <div className="dash-card h-full p-4 flex items-center justify-between">
      <div>
        <p
          className="text-xs font-semibold uppercase tracking-wider mb-1"
          style={{ color: 'var(--dash-text-dim)' }}
        >
          {label}
        </p>
        {isLoading ? (
          <Skeleton className="h-7 w-20 my-1" />
        ) : (
          <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
            {value}
          </p>
        )}
        <p className="text-[11px] mt-1" style={{ color: 'var(--dash-text-dim)' }}>
          {sub}
        </p>
      </div>
      <div
        className="h-10 w-10 rounded-xl flex items-center justify-center"
        style={{ background: `${color}15`, color }}
      >
        <Icon size={18} />
      </div>
    </div>
  );
}

export const KpiRevenue: React.FC = () => {
  const { data: kpis, isLoading } = useKpis();
  return (
    <KpiCard
      label="Revenue"
      value={fmt(kpis?.revenue)}
      sub=""
      icon={DollarSign}
      color="#4583C4"
      isLoading={isLoading}
    />
  );
};

export const KpiOrders: React.FC = () => {
  const { data: kpis, isLoading } = useKpis();
  return (
    <KpiCard
      label="Orders"
      value={kpis?.orderCount ?? 0}
      sub=""
      icon={ShoppingBag}
      color="#4583C4"
      isLoading={isLoading}
    />
  );
};

export const KpiTeam: React.FC = () => {
  const { data: kpis, isLoading } = useKpis();
  return (
    <KpiCard
      label="Team"
      value={kpis?.staffCount ?? 0}
      sub="staff"
      icon={Users}
      color="#f59e0b"
      isLoading={isLoading}
    />
  );
};

export const KpiInvoices: React.FC = () => {
  // Not from useKpis: hub.dashboard.v1 has no invoice count, so this card read a
  // field that did not exist and showed a confident 0 forever (rtylr#1001).
  // insights.financial.v1 getInvoiceAging already carries it, and dash already
  // calls that endpoint for InvoiceAgingWidget.
  const { data: openInvoices, isLoading } = useOpenInvoiceCount();
  return (
    <KpiCard
      label="Invoices"
      value={openInvoices ?? 0}
      sub="open"
      icon={FileText}
      color="#10b981"
      isLoading={isLoading}
    />
  );
};
