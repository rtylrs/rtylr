import React from 'react';
import { ShoppingBag } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { toFiniteNumber, unwrap } from '@voxire/api';
import { safe } from '../../../../shared/lib/api';

interface WidgetProduct {
  productId?: number;
  name?: string;
  category?: string;
  unitsSold?: number;
  revenueMinorUnits?: number;
  revenue?: number;
}

interface RawWidgetProduct {
  productId?: number | string;
  product_id?: number | string;
  name?: string;
  category?: string;
  unitsSold?: number | string;
  units_sold?: number | string;
  revenueMinorUnits?: number | string;
  revenue_minor_units?: number | string;
  revenue?: number | string;
}

interface BestSellersResponse {
  products?: RawWidgetProduct[];
}

function productRevenue(product: WidgetProduct): number {
  if (typeof product.revenueMinorUnits === 'number') {
    return product.revenueMinorUnits / 100;
  }
  return product.revenue ?? 0;
}

const TopProductsWidget: React.FC = () => {
  const { data: raw } = useQuery({
    queryKey: ['top-products-widget'],
    queryFn: async (): Promise<WidgetProduct[]> => {
      try {
        const res = await unwrap(
          safe.get<BestSellersResponse | RawWidgetProduct[]>('/insights/product/v1/best-sellers', {
            params: { limit: 5 },
          })
        );
        return ((Array.isArray(res) ? res : res.products) ?? []).map((product) => ({
          productId: toFiniteNumber(product.productId ?? product.product_id),
          name: product.name ?? '',
          category: product.category ?? '',
          unitsSold: toFiniteNumber(product.unitsSold ?? product.units_sold),
          revenueMinorUnits:
            product.revenueMinorUnits !== undefined || product.revenue_minor_units !== undefined
              ? toFiniteNumber(product.revenueMinorUnits ?? product.revenue_minor_units)
              : undefined,
          revenue: product.revenue !== undefined ? toFiniteNumber(product.revenue) : undefined,
        }));
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });
  const products: WidgetProduct[] = Array.isArray(raw) ? (raw as WidgetProduct[]).slice(0, 5) : [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <ShoppingBag size={14} style={{ color: '#8b5cf6' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Top products
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {products.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              No products yet
            </p>
          </div>
        ) : (
          products.map((p, i) => (
            <div key={i} className="flex items-center gap-3 px-2 py-1.5">
              <span
                className="text-[10px] font-bold w-4 text-center"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {i + 1}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--dash-text)' }}>
                  {p.name}
                </p>
                <p className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
                  {p.category || `${p.unitsSold ?? 0} sold`}
                </p>
              </div>
              <span className="text-xs font-semibold" style={{ color: 'var(--dash-text)' }}>
                ${productRevenue(p).toFixed(2)}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default TopProductsWidget;
