import React from 'react';
import { ShoppingBag } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import api from '../../../../shared/lib/api';

interface WidgetProduct {
  name?: string;
  sku?: string;
  category?: string;
  salePrice?: number;
  price?: number;
}

const TopProductsWidget: React.FC = () => {
  const { data: raw } = useQuery({
    queryKey: ['top-products-widget'],
    queryFn: async () => {
      const res = await api.get('/erp/products?limit=5');
      return res.data?.products ?? res.data ?? [];
    },
    staleTime: 60_000,
  });
  const products: WidgetProduct[] = Array.isArray(raw) ? (raw as WidgetProduct[]).slice(0, 5) : [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--hub-border)' }}
      >
        <ShoppingBag size={14} style={{ color: '#8b5cf6' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
          Top products
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {products.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              No products yet
            </p>
          </div>
        ) : (
          products.map((p, i) => (
            <div key={i} className="flex items-center gap-3 px-2 py-1.5">
              <span
                className="text-[10px] font-bold w-4 text-center"
                style={{ color: 'var(--hub-text-dim)' }}
              >
                {i + 1}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                  {p.name}
                </p>
                <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                  {p.sku || p.category || ''}
                </p>
              </div>
              <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                ${p.salePrice ?? p.price ?? 0}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default TopProductsWidget;
