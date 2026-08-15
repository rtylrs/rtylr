import React from 'react';
import { motion } from 'motion/react';
import { Sparkles, TrendingUp } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useReorderSuggestions } from './useReorderSuggestions';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';

const AMBER = 'var(--erp-accent)';

const Reorder: React.FC = () => {
  const { data, isLoading } = useReorderSuggestions();
  const suggestions = data?.suggestions ?? [];

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Reorder suggestions · rtylr" />
      <PageHeader
        title="Reorder Suggestions"
        description="Smart recommendations based on sales velocity and lead times"
      />

      <div
        className="rounded-xl p-4 flex items-center gap-3"
        style={{
          background: 'color-mix(in srgb, var(--erp-accent) 8%, transparent)',
          border: '1px solid color-mix(in srgb, var(--erp-accent) 20%, transparent)',
        }}
      >
        <Sparkles size={16} style={{ color: AMBER }} />
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Suggestions use: current stock / avg daily sales / supplier lead time
        </p>
      </div>

      {data?.truncated && (
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Showing the 100 most urgent suggestions. Refine inventory thresholds to reduce this
          preview.
        </p>
      )}

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3 text-[10px] font-semibold uppercase tracking-wider"
          style={{
            gridTemplateColumns: '2fr 0.8fr 1fr 1fr 0.8fr 0.8fr',
            color: 'var(--hub-text-dim)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <span>Product</span>
          <span>Stock</span>
          <span>Avg Daily</span>
          <span>Suggested</span>
          <span>Lead</span>
          <span>Confidence</span>
        </div>
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={6} />)
        ) : suggestions.length === 0 ? (
          <EmptyState title="No suggestions" description="All stock levels are healthy" />
        ) : (
          suggestions.map((s) => {
            const confColor =
              s.confidence === 'high'
                ? 'var(--erp-success)'
                : s.confidence === 'medium'
                  ? AMBER
                  : 'var(--hub-text-dim)';
            return (
              <div
                key={s.productId}
                className="grid px-4 py-3 items-center"
                style={{
                  gridTemplateColumns: '2fr 0.8fr 1fr 1fr 0.8fr 0.8fr',
                  borderTop: '1px solid var(--hub-border)',
                }}
              >
                <div>
                  <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                    {s.name}
                  </p>
                  <p className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                    {s.sku}
                  </p>
                </div>
                <span
                  className="text-sm font-semibold"
                  style={{
                    color: s.currentStock === 0 ? 'var(--erp-danger)' : 'var(--hub-text)',
                  }}
                >
                  {s.currentStock}
                </span>
                <span
                  className="text-xs flex items-center gap-1"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  <TrendingUp size={11} /> {s.avgDailySales}/day
                </span>
                <span className="text-sm font-bold" style={{ color: AMBER }}>
                  {s.suggestedQuantity} units
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {s.leadTimeDays}d
                </span>
                <span>
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                    style={{
                      background: `color-mix(in srgb, ${confColor} 13%, transparent)`,
                      color: confColor,
                    }}
                  >
                    {s.confidence}
                  </span>
                </span>
              </div>
            );
          })
        )}
      </div>
    </motion.div>
  );
};

export default Reorder;
