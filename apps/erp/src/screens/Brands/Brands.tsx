import React from 'react';
import { motion } from 'framer-motion';
import { Plus, Tag } from 'lucide-react';
import { useBrands } from '../../hooks/useBrands';
import PageHeader from '../../components/PageHeader';
import EmptyState from '../../components/EmptyState';
import { Skeleton } from '../../components/Skeleton';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const Brands: React.FC = () => {
  const { data: brands = [], isLoading } = useBrands();

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
      <PageHeader
        title="Brands"
        description={`${brands.length} brands across your catalog`}
        actions={
          <button
            onClick={() => toast.success('Brand created')}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Brand
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-24 w-full rounded-xl" />
          ))}
        </div>
      ) : brands.length === 0 ? (
        <EmptyState title="No brands yet" description="Add your first brand to group products." />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {brands.map((b) => (
            <div
              key={b.id}
              className="rounded-xl p-4 transition-colors"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'rgba(245,158,11,0.3)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)'; }}
            >
              <div className="flex items-center gap-3 mb-3">
                <div
                  className="h-10 w-10 rounded-lg flex items-center justify-center font-bold text-sm"
                  style={{ background: 'rgba(245,158,11,0.12)', color: AMBER }}
                >
                  {b.name.split(' ').map((w) => w[0]).join('').slice(0, 2)}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold truncate" style={{ color: 'var(--hub-text)' }}>
                    {b.name}
                  </p>
                  <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                    {b.productCount} products
                  </p>
                </div>
                <Tag size={14} style={{ color: 'var(--hub-text-dim)' }} />
              </div>
              {b.description && (
                <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {b.description}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </motion.div>
  );
};

export default Brands;
