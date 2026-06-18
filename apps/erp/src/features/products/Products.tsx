import React, { useMemo, useState } from 'react';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { Search, Plus, Grid3X3, List, SlidersHorizontal, ArrowUpDown, Package } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useDebounce } from '@voxire/utils';
import { useProducts, useDeleteProduct, ProductFilters } from './useProducts';
import { type Product } from '../../shared/services/erpApi';
import { useCategories } from '../categories/useCategories';
import { useBrands } from '../brands/useBrands';
import ProductCard from '../../shared/components/ProductCard';
import StockBadge from '../../shared/components/StockBadge';
import FilterDrawer from '../../shared/components/FilterDrawer';
import BulkActionBar from '../../shared/components/BulkActionBar';
import ExportButton from '../../shared/components/ExportButton';
import ImportCSV from '../../shared/components/ImportCSV';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import QuickAddProductModal from '../../shared/components/QuickAddProductModal';
import PageHeader from '../../shared/components/PageHeader';
import { formatCurrency } from '../../shared/lib/format';
import { toast } from 'sonner';
import { Link } from '@tanstack/react-router';

const AMBER = '#f59e0b';

const Products: React.FC = () => {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const [filters, setFilters] = useState<ProductFilters>({});
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('list');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [showAdd, setShowAdd] = useState(false);

  const mergedFilters: ProductFilters = { ...filters, search: debouncedSearch };
  const { data: productResult, isLoading } = useProducts(mergedFilters);
  const products = productResult?.products ?? [];
  const { data: categories = [] } = useCategories();
  const { data: brands = [] } = useBrands();
  const deleteMut = useDeleteProduct();

  const toggle = (id: number) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelected(next);
  };

  const handleBulkDelete = async (): Promise<void> => {
    const ids = Array.from(selected);
    try {
      await Promise.all(ids.map((id) => deleteMut.mutateAsync(id)));
      toast.success(t('erp.products.toast.deleted'));
      setSelected(new Set());
    } catch {
      // onError in hook handles toast
    }
  };

  const exportable = useMemo(
    () =>
      products.map((p) => ({
        id: p.id,
        name: p.name,
        sku: p.sku,
        category: p.category,
        brand: p.brand || '',
        stock: p.stock,
        costPrice: p.costPrice,
        salePrice: p.salePrice,
        status: p.status,
      })),
    [products]
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Products · rtylr" />
      <PageHeader
        title={t('erp.products.title')}
        description={t('erp.products.catalogCount', { count: products.length })}
        actions={
          <>
            <ImportCSV
              onImport={(rows) =>
                toast.info(`${rows.length} rows parsed - bulk import coming soon`)
              }
              expectedColumns={['name', 'sku', 'category', 'stock']}
            />
            <ExportButton data={exportable} filename="products.csv" />
            <button
              data-testid="products-new-button"
              onClick={() => setShowAdd(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: AMBER, color: '#000' }}
            >
              <Plus size={14} />
              {t('erp.products.new')}
            </button>
          </>
        }
      />

      {/* Toolbar */}
      <div className="flex items-center gap-2 flex-wrap">
        <div className="relative flex-1 max-w-sm">
          <Search
            size={14}
            className="absolute left-3 top-1/2 -translate-y-1/2"
            style={{ color: 'var(--hub-text-dim)' }}
          />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search products…"
            className="w-full pl-9 pr-4 py-2 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </div>
        <button
          onClick={() => setDrawerOpen(true)}
          className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium"
          style={{
            background: 'var(--hub-hover)',
            color: 'var(--hub-text-muted)',
            border: '1px solid var(--hub-border)',
          }}
        >
          <SlidersHorizontal size={12} />
          Filters
          {(filters.categoryId !== undefined || filters.brand || filters.status) && (
            <span className="h-1.5 w-1.5 rounded-full" style={{ background: AMBER }} />
          )}
        </button>
        <button
          onClick={() =>
            setFilters({
              ...filters,
              sort: 'name',
              sortDir: filters.sortDir === 'asc' ? 'desc' : 'asc',
            })
          }
          className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium"
          style={{
            background: 'var(--hub-hover)',
            color: 'var(--hub-text-muted)',
            border: '1px solid var(--hub-border)',
          }}
        >
          <ArrowUpDown size={12} />
          Sort
        </button>
        <div
          className="flex rounded-lg p-0.5"
          style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
        >
          <button
            onClick={() => setViewMode('list')}
            className="h-7 w-7 rounded-md flex items-center justify-center"
            style={{
              background: viewMode === 'list' ? 'var(--hub-active)' : 'transparent',
              color: viewMode === 'list' ? 'var(--hub-text)' : 'var(--hub-text-dim)',
            }}
          >
            <List size={12} />
          </button>
          <button
            onClick={() => setViewMode('grid')}
            className="h-7 w-7 rounded-md flex items-center justify-center"
            style={{
              background: viewMode === 'grid' ? 'var(--hub-active)' : 'transparent',
              color: viewMode === 'grid' ? 'var(--hub-text)' : 'var(--hub-text-dim)',
            }}
          >
            <Grid3X3 size={12} />
          </button>
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 6 }).map((_, i) => (
            <SkeletonRow key={i} cols={6} />
          ))}
        </div>
      ) : products.length === 0 ? (
        <EmptyState
          title={t('erp.products.empty.title')}
          description={t('erp.products.empty.description')}
          action={
            <button
              data-testid="products-empty-create-button"
              onClick={() => setShowAdd(true)}
              className="px-3 py-2 rounded-lg text-xs font-semibold"
              style={{ background: AMBER, color: '#000' }}
            >
              {t('erp.products.create')}
            </button>
          }
        />
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          {products.map((p) => (
            <ProductCard key={p.id} product={p} selected={selected.has(p.id)} onToggle={toggle} />
          ))}
        </div>
      ) : (
        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div
            className="grid px-4 py-3 text-[10px] font-semibold uppercase tracking-wider"
            style={{
              gridTemplateColumns: '0.2fr 2fr 1fr 1fr 1fr 1fr 1fr 0.8fr',
              color: 'var(--hub-text-dim)',
              borderBottom: '1px solid var(--hub-border)',
            }}
          >
            <span></span>
            <span>Name</span>
            <span>SKU</span>
            <span>Category</span>
            <span>Brand</span>
            <span>Stock</span>
            <span>Price</span>
            <span>Status</span>
          </div>
          <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
            {products.map((p) => (
              <div
                key={p.id}
                className="grid px-4 py-3 items-center transition-colors"
                style={{
                  gridTemplateColumns: '0.2fr 2fr 1fr 1fr 1fr 1fr 1fr 0.8fr',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'transparent';
                }}
              >
                <input
                  type="checkbox"
                  checked={selected.has(p.id)}
                  onChange={() => toggle(p.id)}
                  className="cursor-pointer"
                />
                <div className="flex min-w-0 items-center gap-3">
                  <div
                    className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg"
                    style={{ background: 'var(--hub-hover)' }}
                  >
                    {p.imageUrl ? (
                      <img
                        src={p.imageUrl}
                        alt={p.name}
                        className="h-full w-full bg-white p-1 object-contain"
                        loading="lazy"
                      />
                    ) : (
                      <Package size={16} style={{ color: 'var(--hub-text-dim)' }} />
                    )}
                  </div>
                  <Link
                    to="/products/$id"
                    params={{ id: String(p.id) }}
                    className="min-w-0 truncate text-sm font-medium hover:underline"
                    style={{ color: 'var(--hub-text)', textDecoration: 'none' }}
                  >
                    {p.name}
                  </Link>
                </div>
                <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                  {p.sku}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {p.category}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {p.brand || '-'}
                </span>
                <span
                  className="text-sm font-semibold"
                  style={{
                    color:
                      p.stock === 0 ? '#f87171' : p.stock < p.lowStock ? AMBER : 'var(--hub-text)',
                  }}
                >
                  {p.stock}
                </span>
                <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(p.salePrice)}
                </span>
                <StockBadge stock={p.stock} lowStock={p.lowStock} />
              </div>
            ))}
          </div>
        </div>
      )}

      <FilterDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        onReset={() => setFilters({})}
      >
        <div>
          <label
            className="text-xs font-semibold mb-1 block"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            Category
          </label>
          <select
            value={filters.categoryId !== undefined ? String(filters.categoryId) : ''}
            onChange={(e) =>
              setFilters({
                ...filters,
                categoryId: e.target.value ? Number(e.target.value) : undefined,
              })
            }
            className="w-full px-2 py-1.5 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            <option value="">All categories</option>
            {categories.map((c) => (
              <option key={c.id} value={String(c.id)}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label
            className="text-xs font-semibold mb-1 block"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            Brand
          </label>
          <select
            value={filters.brand || ''}
            onChange={(e) => setFilters({ ...filters, brand: e.target.value || undefined })}
            className="w-full px-2 py-1.5 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            <option value="">All brands</option>
            {brands.map((b) => (
              <option key={b.id} value={b.name}>
                {b.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label
            className="text-xs font-semibold mb-1 block"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            Status
          </label>
          <select
            value={filters.status || ''}
            onChange={(e) =>
              setFilters({ ...filters, status: (e.target.value as Product['status']) || undefined })
            }
            className="w-full px-2 py-1.5 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            <option value="">All statuses</option>
            <option value="In Stock">In Stock</option>
            <option value="Low Stock">Low Stock</option>
            <option value="Out of Stock">Out of Stock</option>
          </select>
        </div>
      </FilterDrawer>

      <BulkActionBar
        count={selected.size}
        onClear={() => setSelected(new Set())}
        onDelete={handleBulkDelete}
        onArchive={() => {
          toast.info('Archive coming soon');
        }}
        onExport={() => {
          const rows = products.filter((p) => selected.has(p.id));
          if (rows.length === 0) return;
          const header = [
            'id',
            'name',
            'sku',
            'category',
            'brand',
            'stock',
            'costPrice',
            'salePrice',
            'status',
          ];
          const csv = [
            header,
            ...rows.map((p) => [
              p.id,
              p.name,
              p.sku,
              p.category,
              p.brand || '',
              p.stock,
              p.costPrice,
              p.salePrice,
              p.status,
            ]),
          ]
            .map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(','))
            .join('\n');
          const a = document.createElement('a');
          a.href = URL.createObjectURL(new Blob([csv], { type: 'text/csv' }));
          a.download = `products-selection.csv`;
          a.click();
          toast.success(`Exported ${rows.length} products`);
        }}
      />

      <QuickAddProductModal open={showAdd} onClose={() => setShowAdd(false)} />
    </motion.div>
  );
};

export default Products;
