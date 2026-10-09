import React, { useState } from 'react';
import { motion, type Variants } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { cdnImageUrl } from '@voxire/config';
import { Package, Search, SearchX, Plus } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { useDebounce } from '@voxire/utils';
import {
  getProductStockStatus,
  type ProductStockStatus,
  useProducts,
} from '../products/useProducts';
import QuickAddProductModal from '../../shared/components/QuickAddProductModal';
import PaginationControls from '../../shared/components/PaginationControls';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const ACCENT = 'var(--erp-accent)';

// ─── Config ───────────────────────────────────────────────────────────────────

const statusConfig: Record<ProductStockStatus, { color: string; bg: string }> = {
  'In Stock': {
    color: 'var(--erp-success)',
    bg: 'color-mix(in srgb, var(--erp-success) 10%, transparent)',
  },
  'Low Stock': {
    color: 'var(--erp-warning)',
    bg: 'color-mix(in srgb, var(--erp-warning) 10%, transparent)',
  },
  'Out of Stock': {
    color: 'var(--erp-danger)',
    bg: 'color-mix(in srgb, var(--erp-danger) 10%, transparent)',
  },
};

// ─── Animation variants ───────────────────────────────────────────────────────

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const columns = ['Product', 'SKU', 'Stock', 'Low Stock', 'Cost Price', 'Sale Price', 'Status'];

const tableColumns = 'minmax(260px, 2.2fr) minmax(140px, 1fr) 0.7fr 0.7fr 0.9fr 0.9fr 0.9fr';

// ─── Skeleton row ─────────────────────────────────────────────────────────────

const SkeletonTableRow: React.FC = () => (
  <div className="grid px-4 py-3 items-center" style={{ gridTemplateColumns: tableColumns }}>
    {Array.from({ length: columns.length }).map((_, i) => (
      <div
        key={i}
        className="h-3 rounded animate-pulse mr-2"
        style={{ background: 'var(--hub-border)', opacity: 0.6 }}
      />
    ))}
  </div>
);

// ─── Component ────────────────────────────────────────────────────────────────

const Inventory: React.FC = () => {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');
  const [quickAddOpen, setQuickAddOpen] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = 50;
  const debouncedSearch = useDebounce(search, 300);
  const { data: productResult, isLoading } = useProducts({
    search: debouncedSearch,
    page,
    pageSize,
  });
  const products = productResult?.products ?? [];
  const totalProducts = productResult?.total ?? products.length;
  const hasMore = productResult?.hasMore ?? false;

  const changePage = (nextPage: number): void => {
    setPage(nextPage);
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/inventory'].title))} />
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('erp.inventory.title')}
        </h2>
        <button
          type="button"
          data-testid="inventory-add-product-button"
          onClick={() => setQuickAddOpen(true)}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLElement).style.opacity = '0.9';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLElement).style.opacity = '1';
          }}
        >
          <Plus size={14} />
          {t('erp.inventory.addProduct')}
        </button>
      </motion.div>

      {/* Search */}
      <motion.div variants={itemVariants} className="relative max-w-sm">
        <Search
          size={14}
          className="absolute left-3 top-1/2 -translate-y-1/2"
          style={{ color: 'var(--hub-text-dim)' }}
        />
        <input
          type="text"
          placeholder={t('erp.products.searchPlaceholder')}
          value={search}
          onChange={(e) => {
            setPage(1);
            setSearch(e.target.value);
          }}
          className="w-full pl-9 pr-4 py-2 rounded-lg text-sm outline-none transition-colors"
          style={{
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text)',
          }}
          onFocus={(e) => {
            (e.currentTarget as HTMLElement).style.borderColor =
              'color-mix(in srgb, var(--erp-accent) 50%, transparent)';
          }}
          onBlur={(e) => {
            (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
          }}
        />
      </motion.div>

      {/* Table */}
      <motion.div
        variants={itemVariants}
        className="overflow-hidden rounded-xl"
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
        }}
      >
        <div className="overflow-x-auto">
          <div className="min-w-[860px]">
            {/* Column headers */}
            <div
              className="grid px-4 py-3"
              style={{
                gridTemplateColumns: tableColumns,
                borderBottom: '1px solid var(--hub-border)',
              }}
            >
              {columns.map((col) => (
                <span
                  key={col}
                  className="text-[10px] font-semibold uppercase tracking-wider"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  {col}
                </span>
              ))}
            </div>

            {/* Rows */}
            <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
              {isLoading ? (
                Array.from({ length: 6 }).map((_, i) => <SkeletonTableRow key={i} />)
              ) : products.length === 0 ? (
                <EmptyState
                  icon={SearchX}
                  title={t('erp.inventory.empty')}
                  action={{
                    label: t('erp.inventory.clearSearch'),
                    onClick: () => {
                      setPage(1);
                      setSearch('');
                    },
                  }}
                />
              ) : (
                products.map((product) => {
                  const status = getProductStockStatus(product);
                  const sc = statusConfig[status];
                  const imageUrl = cdnImageUrl(
                    (product as typeof product & { imageUrl?: string }).imageUrl
                  );
                  return (
                    <div
                      key={product.id}
                      className="grid items-center px-4 py-3 transition-colors"
                      style={{ gridTemplateColumns: tableColumns }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                    >
                      <div className="flex min-w-0 items-center gap-3">
                        <div
                          className="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-lg"
                          style={{ background: 'var(--hub-hover)' }}
                        >
                          {imageUrl ? (
                            <img
                              src={imageUrl}
                              alt={product.name}
                              className="h-full w-full bg-white p-1 object-contain"
                              loading="lazy"
                            />
                          ) : (
                            <Package size={16} style={{ color: 'var(--hub-text-dim)' }} />
                          )}
                        </div>
                        <div className="min-w-0">
                          <div
                            className="truncate text-sm font-semibold"
                            style={{ color: 'var(--hub-text)' }}
                          >
                            {product.name}
                          </div>
                          {product.unit && (
                            <div className="text-[11px]" style={{ color: 'var(--hub-text-muted)' }}>
                              Unit: {product.unit}
                            </div>
                          )}
                        </div>
                      </div>
                      <span
                        className="truncate text-xs font-mono"
                        style={{ color: 'var(--hub-text-muted)' }}
                      >
                        {product.sku}
                      </span>
                      <span
                        className="text-sm font-semibold"
                        style={{
                          color:
                            product.stockQuantity === 0
                              ? 'var(--erp-danger)'
                              : (product.stockQuantity ?? 0) < (product.lowStockThreshold ?? 0)
                                ? 'var(--erp-warning)'
                                : 'var(--hub-text)',
                        }}
                      >
                        {product.stockQuantity}
                      </span>
                      <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                        {product.lowStockThreshold}
                      </span>
                      <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                        ${(product.costPrice ?? 0).toFixed(2)}
                      </span>
                      <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                        ${(product.salePrice ?? 0).toFixed(2)}
                      </span>
                      <span>
                        <span
                          className="rounded-full px-2 py-0.5 text-[10px] font-medium"
                          style={{ background: sc.bg, color: sc.color }}
                        >
                          {status}
                        </span>
                      </span>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>
      </motion.div>
      <PaginationControls
        page={page}
        hasMore={hasMore}
        itemCount={products.length}
        pageSize={pageSize}
        total={totalProducts}
        onPageChange={changePage}
      />
      <QuickAddProductModal open={quickAddOpen} onClose={() => setQuickAddOpen(false)} />
    </motion.div>
  );
};

export default Inventory;
