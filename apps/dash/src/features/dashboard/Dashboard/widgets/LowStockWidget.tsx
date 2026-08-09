import React from 'react';
import { Package, AlertTriangle } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { toFiniteNumber, unwrap } from '@voxire/api';
import { safe } from '../../../../shared/lib/api';

interface StockItem {
  productName?: string;
  product_name?: string;
  name?: string;
  stockQuantity?: number | string;
  stock_quantity?: number | string;
  stock?: number | string;
}

interface LowStockResponse {
  items?: StockItem[];
  products?: StockItem[];
}

const LowStockWidget: React.FC = () => {
  const { data: raw } = useQuery({
    queryKey: ['low-stock-widget'],
    queryFn: async () => {
      const res = await unwrap(
        safe.get<LowStockResponse | StockItem[]>('/erp/inventory/v1/low-stock')
      );
      return (Array.isArray(res) ? res : (res.items ?? res.products)) ?? [];
    },
    staleTime: 60_000,
  });
  const items: StockItem[] = Array.isArray(raw) ? (raw as StockItem[]).slice(0, 6) : [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <Package size={14} style={{ color: 'var(--aging-critical)' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Low stock
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {items.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              All stock levels are healthy
            </p>
          </div>
        ) : (
          items.map((p, i) => (
            <div
              key={i}
              className="flex items-center gap-3 px-2 py-1.5 rounded-lg"
              style={{ background: 'var(--dash-bg)' }}
            >
              <AlertTriangle size={12} style={{ color: 'var(--aging-critical)' }} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--dash-text)' }}>
                  {p.productName ?? p.product_name ?? p.name}
                </p>
              </div>
              <span
                className="text-[10px] font-semibold"
                style={{ color: 'var(--aging-critical)' }}
              >
                {toFiniteNumber(p.stockQuantity ?? p.stock_quantity ?? p.stock)} left
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default LowStockWidget;
