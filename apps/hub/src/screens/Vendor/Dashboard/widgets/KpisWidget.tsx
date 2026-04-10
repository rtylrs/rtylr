import React from 'react';
import { DollarSign, ShoppingBag, Users, TrendingUp } from 'lucide-react';
import { useKpis } from '../../../../hooks/queries/useKpis';

function fmt(n: number | undefined | null) {
  if (n == null) return '$0';
  return `$${n.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
}

const cards = [
  { key: 'revenue', label: 'Revenue today', icon: DollarSign, color: '#4583C4', getValue: (k: any) => fmt(k?.revenueToday), sub: 'vs yesterday' },
  { key: 'orders', label: 'Orders today', icon: ShoppingBag, color: '#8b5cf6', getValue: (k: any) => k?.ordersToday ?? 0, sub: 'vs yesterday' },
  { key: 'team', label: 'Team', icon: Users, color: '#f59e0b', getValue: (k: any) => `${k?.activeStaff ?? 0}/${k?.headcount ?? 0}`, sub: 'on shift' },
  { key: 'invoices', label: 'Open invoices', icon: TrendingUp, color: '#10b981', getValue: (k: any) => k?.openInvoices ?? 0, sub: 'unpaid' },
];

const KpisWidget: React.FC = () => {
  const { data: kpis } = useKpis();

  return (
    <div className="grid grid-cols-4 gap-3 h-full">
      {cards.map((c) => {
        const Icon = c.icon;
        return (
          <div key={c.key} className="rounded-xl p-4" style={{ background: 'var(--hub-bg)', border: '1px solid var(--hub-border)' }}>
            <div className="flex items-center justify-between mb-3">
              <p className="text-[11px] font-medium uppercase tracking-wider" style={{ color: 'var(--hub-text-dim)' }}>{c.label}</p>
              <div className="h-7 w-7 rounded-lg flex items-center justify-center" style={{ background: `${c.color}18` }}>
                <Icon size={14} style={{ color: c.color }} />
              </div>
            </div>
            <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>{c.getValue(kpis)}</p>
            <p className="text-[11px] mt-1" style={{ color: 'var(--hub-text-dim)' }}>{c.sub}</p>
          </div>
        );
      })}
    </div>
  );
};

export default KpisWidget;
