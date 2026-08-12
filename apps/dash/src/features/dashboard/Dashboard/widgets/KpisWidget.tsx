import React from 'react';
import { DollarSign, ShoppingBag, Users, TrendingUp } from 'lucide-react';
import { useKpis } from '../../../../shared/hooks/queries/useKpis';

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
}: {
  label: string;
  value: React.ReactNode;
  sub: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  color: string;
}) {
  return (
    <div className="flex flex-col justify-between h-full p-4">
      <div className="flex items-center justify-between mb-2">
        <p
          className="text-[11px] font-medium uppercase tracking-wider"
          style={{ color: 'var(--dash-text-dim)' }}
        >
          {label}
        </p>
        <div
          className="h-7 w-7 rounded-lg flex items-center justify-center"
          style={{ background: `${color}15` }}
        >
          <Icon size={14} style={{ color }} />
        </div>
      </div>
      <div>
        <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
          {value}
        </p>
        <p className="text-[11px] mt-1" style={{ color: 'var(--dash-text-dim)' }}>
          {sub}
        </p>
      </div>
    </div>
  );
}

export const KpiRevenue: React.FC = () => {
  const { data: kpis } = useKpis();
  return (
    <KpiCard label="Revenue" value={fmt(kpis?.revenue)} sub="" icon={DollarSign} color="#4583C4" />
  );
};

export const KpiOrders: React.FC = () => {
  const { data: kpis } = useKpis();
  return (
    <KpiCard
      label="Orders"
      value={kpis?.orderCount ?? 0}
      sub=""
      icon={ShoppingBag}
      color="#8b5cf6"
    />
  );
};

export const KpiTeam: React.FC = () => {
  const { data: kpis } = useKpis();
  return (
    <KpiCard label="Team" value={kpis?.staffCount ?? 0} sub="staff" icon={Users} color="#f59e0b" />
  );
};

// Widget key stays 'kpi-invoices' (registry/layout key); shows Customers, since
// the backend KPI response has no open-invoices figure.
export const KpiInvoices: React.FC = () => {
  const { data: kpis } = useKpis();
  return (
    <KpiCard
      label="Customers"
      value={kpis?.customerCount ?? 0}
      sub="total"
      icon={TrendingUp}
      color="#10b981"
    />
  );
};
