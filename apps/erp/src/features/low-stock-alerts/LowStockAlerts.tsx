import React from 'react';
import { motion } from 'motion/react';
import { AlertTriangle, RefreshCw, ShoppingCart } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useLowStock } from '../inventory/useInventory';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { toast } from 'sonner';

const AMBER = 'var(--erp-accent)';

const LowStockAlerts: React.FC = () => {
  const { data: items = [], isLoading } = useLowStock();

  const handleReorder = (name: string) => {
    toast.info(`Reorder for ${name} coming soon`);
  };

  const handleReorderAll = () => {
    toast.info('Bulk reorder coming soon');
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Low stock alerts · rtylr" />
      <PageHeader
        title="Low Stock Alerts"
        description={`${items.length} items below minimum stock levels`}
        actions={
          <button
            onClick={handleReorderAll}
            disabled={items.length === 0}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{
              background: AMBER,
              color: 'var(--erp-accent-foreground)',
              opacity: items.length === 0 ? 0.5 : 1,
            }}
          >
            <RefreshCw size={14} /> Reorder All
          </button>
        }
      />

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <SkeletonRow key={i} cols={5} />
          ))}
        </div>
      ) : items.length === 0 ? (
        <EmptyState
          icon={AlertTriangle}
          title="All stock levels healthy"
          description="Nothing needs reordering right now."
        />
      ) : (
        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          {items.map((item, i) => {
            const critical = (item.stockQuantity ?? 0) <= 5;
            return (
              <div
                key={item.productId}
                className="flex items-center gap-3 px-4 py-3"
                style={{
                  borderBottom: i < items.length - 1 ? '1px solid var(--hub-border)' : 'none',
                }}
              >
                <div
                  className="h-9 w-9 rounded-lg flex items-center justify-center"
                  style={{
                    background: critical
                      ? 'color-mix(in srgb, var(--erp-danger) 12%, transparent)'
                      : 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
                  }}
                >
                  <AlertTriangle
                    size={14}
                    style={{ color: critical ? 'var(--erp-danger)' : AMBER }}
                  />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                    {item.productName}
                  </p>
                  <p className="text-[11px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                    {item.sku}
                  </p>
                </div>
                <div className="text-right pr-3">
                  <p
                    className="text-sm font-bold"
                    style={{ color: critical ? 'var(--erp-danger)' : AMBER }}
                  >
                    {item.stockQuantity} {item.unit}
                  </p>
                  <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                    Min: {item.lowStockThreshold}
                  </p>
                </div>
                <button
                  onClick={() => handleReorder(item.productName ?? '')}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold"
                  style={{
                    background: 'color-mix(in srgb, var(--erp-accent) 15%, transparent)',
                    color: AMBER,
                  }}
                >
                  <ShoppingCart size={12} /> Reorder
                </button>
              </div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default LowStockAlerts;
