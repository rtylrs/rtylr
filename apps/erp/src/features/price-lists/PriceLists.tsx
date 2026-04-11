import React from 'react';
import { motion } from 'motion/react';
import { Plus, Tags } from 'lucide-react';
import { usePriceLists } from './usePriceLists';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const tierConfig = {
  Retail: { color: '#10b981', bg: 'rgba(16,185,129,0.12)' },
  Wholesale: { color: AMBER, bg: 'rgba(245,158,11,0.12)' },
  VIP: { color: '#a78bfa', bg: 'rgba(167,139,250,0.12)' },
  Custom: { color: '#60a5fa', bg: 'rgba(96,165,250,0.12)' },
};

const PriceLists: React.FC = () => {
  const { data: lists = [], isLoading } = usePriceLists();

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
      <PageHeader
        title="Price Lists"
        description="Customer-tier pricing strategies"
        actions={
          <button
            onClick={() => toast.success('Price list created')}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Price List
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-32 rounded-xl" />)}
        </div>
      ) : lists.length === 0 ? (
        <EmptyState icon={Tags} title="No price lists" description="Create your first price list to enable tier pricing" />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {lists.map((l) => {
            const cfg = tierConfig[l.tier];
            return (
              <div
                key={l.id}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div
                      className="h-10 w-10 rounded-lg flex items-center justify-center"
                      style={{ background: cfg.bg }}
                    >
                      <Tags size={16} style={{ color: cfg.color }} />
                    </div>
                    <div>
                      <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                        {l.name}
                      </p>
                      <span
                        className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                        style={{ background: cfg.bg, color: cfg.color }}
                      >
                        {l.tier}
                      </span>
                    </div>
                  </div>
                  <span className="text-xl font-bold" style={{ color: cfg.color }}>
                    {Math.round(l.multiplier * 100)}%
                  </span>
                </div>
                <p className="text-xs mb-3" style={{ color: 'var(--hub-text-muted)' }}>
                  {l.description}
                </p>
                <div
                  className="flex items-center justify-between pt-3"
                  style={{ borderTop: '1px solid var(--hub-border)' }}
                >
                  <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                    Active products
                  </span>
                  <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {l.activeCount}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default PriceLists;
