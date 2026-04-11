import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Package, Search, Grid3X3, List, Plus } from 'lucide-react';
import { useDebounce } from '@voxire/utils';
import { useProducts, ProductFilters } from './useProducts';
import { formatCurrency } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';

const TEAL = '#14b8a6';

function getStatusStyle(status: string): { bg: string; color: string } {
  if (status === 'active') return { bg: 'rgba(16,185,129,0.1)', color: '#10b981' };
  if (status === 'draft') return { bg: 'rgba(245,158,11,0.1)', color: '#f59e0b' };
  return { bg: 'rgba(107,114,128,0.1)', color: '#6b7280' };
}

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.03 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.18, ease: 'easeOut' } },
};

const Products: React.FC = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const [statusFilter, setStatusFilter] = useState<ProductFilters['status']>();
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('list');

  const { data: products, isLoading, error } = useProducts({
    search: debouncedSearch || undefined,
    status: statusFilter,
  });

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>Failed to load products</p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>Please check your connection and try again.</p>
      </div>
    );
  }

  return (
    <motion.div variants={containerVariants} initial="hidden" animate="show" className="space-y-4 max-w-7xl">
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Package size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Products</h2>
          {!isLoading && (
            <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}>
              {(products ?? []).length}
            </span>
          )}
        </div>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: TEAL, color: '#fff' }}
        >
          <Plus size={14} />
          Add Product
        </button>
      </motion.div>

      {/* Filters */}
      <motion.div variants={itemVariants} className="flex items-center gap-3">
        <div
          className="flex items-center gap-2 flex-1 max-w-sm px-3 py-2 rounded-lg"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={13} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent text-sm outline-none"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <select
          value={statusFilter ?? ''}
          onChange={(e) => setStatusFilter((e.target.value || undefined) as ProductFilters['status'])}
          className="px-3 py-2 rounded-lg text-sm"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
        >
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="draft">Draft</option>
          <option value="archived">Archived</option>
        </select>
        <div className="flex gap-1">
          <button
            onClick={() => setViewMode('list')}
            className="h-8 w-8 rounded-lg flex items-center justify-center"
            style={{ background: viewMode === 'list' ? 'var(--hub-active)' : 'transparent', color: 'var(--hub-text-muted)' }}
          >
            <List size={14} />
          </button>
          <button
            onClick={() => setViewMode('grid')}
            className="h-8 w-8 rounded-lg flex items-center justify-center"
            style={{ background: viewMode === 'grid' ? 'var(--hub-active)' : 'transparent', color: 'var(--hub-text-muted)' }}
          >
            <Grid3X3 size={14} />
          </button>
        </div>
      </motion.div>

      {/* Product list */}
      {isLoading ? (
        <div className="space-y-1">
          {Array.from({ length: 8 }).map((_, i) => <SkeletonRow key={i} cols={5} />)}
        </div>
      ) : (products ?? []).length === 0 ? (
        <motion.div variants={itemVariants} className="flex flex-col items-center py-16 gap-3">
          <Package size={32} style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>No products found</p>
        </motion.div>
      ) : viewMode === 'list' ? (
        <motion.div variants={containerVariants} className="space-y-1">
          <div className="grid grid-cols-6 gap-2 px-3 py-2 text-[10px] uppercase tracking-wider font-semibold"
            style={{ color: 'var(--hub-text-dim)' }}>
            <span className="col-span-2">Product</span>
            <span>SKU</span>
            <span>Price</span>
            <span>Stock</span>
            <span>Status</span>
          </div>
          {(products ?? []).map((product) => {
            const statusStyle = getStatusStyle(product.status);
            return (
              <motion.div
                key={product.id}
                variants={itemVariants}
                className="grid grid-cols-6 gap-2 items-center px-3 py-2.5 rounded-lg cursor-pointer"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div className="col-span-2 flex items-center gap-3 min-w-0">
                  <div
                    className="h-8 w-8 rounded-lg shrink-0 flex items-center justify-center"
                    style={{ background: 'rgba(20,184,166,0.08)' }}
                  >
                    <Package size={12} style={{ color: TEAL }} />
                  </div>
                  <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                    {product.name}
                  </span>
                </div>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{product.sku}</span>
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(product.price)}
                </span>
                <span className="text-xs" style={{ color: product.stock <= 0 ? '#ef4444' : 'var(--hub-text-muted)' }}>
                  {product.stock}
                </span>
                <span
                  className="text-[10px] font-medium px-2 py-0.5 rounded-full capitalize w-fit"
                  style={{ background: statusStyle.bg, color: statusStyle.color }}
                >
                  {product.status}
                </span>
              </motion.div>
            );
          })}
        </motion.div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {(products ?? []).map((product) => {
            const statusStyle = getStatusStyle(product.status);
            return (
              <motion.div
                key={product.id}
                variants={itemVariants}
                className="rounded-xl p-4 cursor-pointer transition-all"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'rgba(20,184,166,0.3)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)'; }}
              >
                <div
                  className="h-24 rounded-lg mb-3 flex items-center justify-center"
                  style={{ background: 'rgba(20,184,166,0.05)' }}
                >
                  <Package size={24} style={{ color: 'var(--hub-text-dim)' }} />
                </div>
                <p className="text-xs font-medium truncate mb-1" style={{ color: 'var(--hub-text)' }}>
                  {product.name}
                </p>
                <div className="flex items-center justify-between">
                  <span className="text-xs font-semibold" style={{ color: TEAL }}>
                    {formatCurrency(product.price)}
                  </span>
                  <span
                    className="text-[9px] font-medium px-1.5 py-0.5 rounded-full capitalize"
                    style={{ background: statusStyle.bg, color: statusStyle.color }}
                  >
                    {product.status}
                  </span>
                </div>
              </motion.div>
            );
          })}
        </motion.div>
      )}
    </motion.div>
  );
};

export default Products;
