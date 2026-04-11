import React from 'react';
import { Package, AlertTriangle } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import api from "../../../../shared/lib/api";

const LowStockWidget: React.FC = () => {
  const { data: raw } = useQuery({
    queryKey: ['low-stock-widget'],
    queryFn: async () => {
      const res = await api.get('/erp/inventory/low-stock');
      return res.data?.products ?? res.data ?? [];
    },
    staleTime: 60_000,
  });
  const items = Array.isArray(raw) ? raw.slice(0, 6) : [];

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3" style={{ borderBottom: '1px solid var(--hub-border)' }}>
        <Package size={14} style={{ color: '#ef4444' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Low stock</p>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {items.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>All stock levels are healthy</p>
          </div>
        ) : (
          items.map((p: any, i: number) => (
            <div key={i} className="flex items-center gap-3 px-2 py-1.5 rounded-lg" style={{ background: 'var(--hub-bg)' }}>
              <AlertTriangle size={12} style={{ color: '#ef4444' }} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>{p.name}</p>
              </div>
              <span className="text-[10px] font-semibold" style={{ color: '#ef4444' }}>{p.stockQuantity ?? p.stock ?? 0} left</span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default LowStockWidget;
