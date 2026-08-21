import React from 'react';
import { useTranslation } from 'react-i18next';
import { Package, AlertTriangle } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { toFiniteNumber } from '@voxire/api';
import { EmptyState } from '@voxire/ui';
import { inventoryClient } from '../../../../shared/lib/clients';
import WidgetHeader from './WidgetHeader';

interface StockItem {
  productName?: string;
  stockQuantity?: number;
}

const LowStockWidget: React.FC = () => {
  const { t } = useTranslation();
  const { data: raw, refetch } = useQuery({
    queryKey: ['low-stock-widget'],
    queryFn: async (): Promise<StockItem[]> => {
      const res = await inventoryClient.getLowStock({});
      return (res.items ?? []).slice(0, 6);
    },
    staleTime: 60_000,
  });
  const items: StockItem[] = raw ?? [];

  return (
    <div className="flex flex-col h-full">
      <WidgetHeader icon={Package} iconColor="var(--aging-critical)" title="Low stock" />
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {items.length === 0 ? (
          <EmptyState
            icon={Package}
            title={t('dashboard.lowStockWidget.emptyTitle')}
            description={t('dashboard.lowStockWidget.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => void refetch() }}
            size="compact"
          />
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
                  {p.productName ?? 'Product'}
                </p>
              </div>
              <span
                className="text-[10px] font-semibold"
                style={{ color: 'var(--aging-critical)' }}
              >
                {toFiniteNumber(p.stockQuantity)} left
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default LowStockWidget;
