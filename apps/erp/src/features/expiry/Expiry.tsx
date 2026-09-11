import React from 'react';
import { motion } from 'motion/react';
import { CalendarX } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { useExpiry } from './useExpiry';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { formatDate, daysUntil } from '../../shared/lib/format';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';
const AMBER = 'var(--erp-accent)';

const Expiry: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data: items = [], isLoading, refetch } = useExpiry();

  const sorted = [...items].sort(
    (a, b) => daysUntil(a.expiryDate ?? '') - daysUntil(b.expiryDate ?? '')
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/expiry'].title))} />
      <PageHeader
        title="Expiry Tracking"
        description={`${items.length} batches with expiry dates`}
      />

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3 text-[10px] font-semibold uppercase tracking-wider"
          style={{
            gridTemplateColumns: '2fr 0.8fr 0.8fr 1.2fr 1fr 0.8fr',
            color: 'var(--hub-text-dim)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <span>Product</span>
          <span>Batch</span>
          <span>Qty</span>
          <span>Warehouse</span>
          <span>Expiry</span>
          <span>Days Left</span>
        </div>
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={6} />)
        ) : sorted.length === 0 ? (
          <EmptyState
            icon={CalendarX}
            title={t('erp.expiry.empty.title')}
            description={t('erp.expiry.empty.description')}
            action={{ label: t('erp.expiry.refresh'), onClick: () => void refetch() }}
          />
        ) : (
          sorted.map((item, i) => {
            const days = daysUntil(item.expiryDate ?? '');
            const isExpired = days < 0;
            const isUrgent = days <= 7 && days >= 0;
            const color = isExpired ? 'var(--erp-danger)' : isUrgent ? AMBER : 'var(--erp-success)';
            return (
              <div
                key={`${item.productId}-${item.batchId}`}
                className="grid px-4 py-3 items-center"
                style={{
                  gridTemplateColumns: '2fr 0.8fr 0.8fr 1.2fr 1fr 0.8fr',
                  borderTop: i > 0 ? '1px solid var(--hub-border)' : 'none',
                }}
              >
                <div>
                  <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                    {item.name}
                  </p>
                  <p className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                    {item.sku}
                  </p>
                </div>
                <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                  {item.batchId}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text)' }}>
                  {item.quantity}
                </span>
                <span className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>
                  {item.warehouseName}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text)' }}>
                  {formatDate(item.expiryDate ?? '')}
                </span>
                <span
                  className="text-[10px] font-semibold px-2 py-0.5 rounded-full w-fit"
                  style={{ background: `color-mix(in srgb, ${color} 13%, transparent)`, color }}
                >
                  {isExpired ? `${Math.abs(days)}d overdue` : `${days}d`}
                </span>
              </div>
            );
          })
        )}
      </div>
    </motion.div>
  );
};

export default Expiry;
