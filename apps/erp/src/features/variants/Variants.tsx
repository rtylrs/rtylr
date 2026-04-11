import React from 'react';
import { motion } from 'motion/react';
import { Plus, Layers } from 'lucide-react';
import { useVariants } from '../products/useVariants';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { formatCurrency } from '../../shared/lib/format';
import StockBadge from '../../shared/components/StockBadge';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const Variants: React.FC = () => {
  const { data: variants = [], isLoading } = useVariants();

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
      <PageHeader
        title="Product Variants"
        description="Individual SKUs for sizes, colors, flavors and more"
        actions={
          <button
            onClick={() => toast.success('Variant created')}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Variant
          </button>
        }
      />

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3 text-[10px] font-semibold uppercase tracking-wider"
          style={{
            gridTemplateColumns: '2fr 1fr 1.5fr 0.8fr 0.8fr 0.8fr',
            color: 'var(--hub-text-dim)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <span>Variant</span>
          <span>SKU</span>
          <span>Attributes</span>
          <span>Stock</span>
          <span>Price</span>
          <span>Status</span>
        </div>
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={6} />)
        ) : variants.length === 0 ? (
          <EmptyState icon={Layers} title="No variants" description="Add variants to differentiate your products" />
        ) : (
          variants.map((v, i) => (
            <div
              key={v.id}
              className="grid px-4 py-3 items-center"
              style={{
                gridTemplateColumns: '2fr 1fr 1.5fr 0.8fr 0.8fr 0.8fr',
                borderTop: i > 0 ? '1px solid var(--hub-border)' : 'none',
              }}
            >
              <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                {v.name}
              </span>
              <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                {v.sku}
              </span>
              <div className="flex gap-1 flex-wrap">
                {Object.entries(v.attributes).map(([k, val]) => (
                  <span
                    key={k}
                    className="text-[10px] px-1.5 py-0.5 rounded"
                    style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                  >
                    {k}: {val}
                  </span>
                ))}
              </div>
              <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                {v.stock}
              </span>
              <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                {formatCurrency(v.price)}
              </span>
              <StockBadge stock={v.stock} lowStock={10} />
            </div>
          ))
        )}
      </div>
    </motion.div>
  );
};

export default Variants;
