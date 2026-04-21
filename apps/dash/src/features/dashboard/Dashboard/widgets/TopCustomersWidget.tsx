import React from 'react';
import { Crown } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import api from '../../../../shared/lib/api';

/**
 * Top 5 customers by lifetime spend. Reads `/insights/customers/top`.
 * Cheap, high-signal widget — restaurants/retailers pull this one
 * open on day 30 to check who their VIPs are.
 */

interface TopCustomer {
  id: number;
  name: string;
  lifetimeSpendMinorUnits: number;
  orderCount: number;
}

interface ApiResponse {
  customers?: {
    id: number;
    name: string;
    lifetime_spend_minor_units: number;
    order_count: number;
  }[];
}

function formatMoney(minor: number): string {
  const major = minor / 100;
  if (major >= 1_000_000) return `$${(major / 1_000_000).toFixed(1)}M`;
  if (major >= 1_000) return `$${(major / 1_000).toFixed(1)}K`;
  return `$${major.toFixed(0)}`;
}

const TopCustomersWidget: React.FC = (): React.JSX.Element => {
  const { data: raw } = useQuery({
    queryKey: ['insights-top-customers', { limit: 5 }],
    queryFn: async (): Promise<TopCustomer[]> => {
      try {
        const res = await api.get<ApiResponse>('/insights/customers/top', {
          params: { limit: 5 },
        });
        return (res.data.customers ?? []).map((c) => ({
          id: c.id,
          name: c.name,
          lifetimeSpendMinorUnits: c.lifetime_spend_minor_units,
          orderCount: c.order_count,
        }));
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });

  const customers = raw ?? [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <Crown size={14} style={{ color: 'var(--accent)' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Top customers
        </p>
      </div>
      {customers.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center text-center p-4">
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            Your VIPs show up here once repeat orders start rolling.
          </p>
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
          {customers.map((c, i) => (
            <div key={c.id} className="flex items-center gap-3 px-2 py-1.5">
              <span
                className="text-[10px] font-bold w-4 text-center"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {i + 1}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--dash-text)' }}>
                  {c.name}
                </p>
                <p className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
                  {c.orderCount} order{c.orderCount === 1 ? '' : 's'}
                </p>
              </div>
              <span className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                {formatMoney(c.lifetimeSpendMinorUnits)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default TopCustomersWidget;
