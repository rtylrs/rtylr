import React from 'react';
import { motion } from 'motion/react';
import { useNavigate } from '@tanstack/react-router';
import { AlertTriangle, CheckCircle2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { useLowStock } from '../inventory/useInventory';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const AMBER = 'var(--erp-accent)';

const LowStockAlerts: React.FC = () => {
  const { t } = useTranslation('erp');
  const navigate = useNavigate();
  const { data: items = [], isLoading } = useLowStock();

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/alerts'].title))} />
      <PageHeader
        title="Low Stock Alerts"
        description={`${items.length} items below minimum stock levels`}
      />

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <SkeletonRow key={i} cols={5} />
          ))}
        </div>
      ) : items.length === 0 ? (
        <EmptyState
          icon={CheckCircle2}
          title={t('erp.lowStockAlerts.empty.title')}
          description={t('erp.lowStockAlerts.empty.description')}
          action={{
            label: t('erp.lowStockAlerts.viewInventory'),
            onClick: () => navigate({ to: '/erp/inventory' }),
          }}
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
              </div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default LowStockAlerts;
