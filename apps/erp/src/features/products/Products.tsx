import React, { useMemo, useState } from 'react';
import { motion } from 'framer-motion';
import { Search, Plus, Grid3X3, List, SlidersHorizontal, ArrowUpDown } from 'lucide-react';
import { useDebounce } from '@voxire/utils';
import { useProducts, ProductFilters } from './useProducts';
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
import { Link } from 'react-router-dom';

const AMBER = '#f59e0b';

const Products: React.FC = () => {
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);
  const [filters, setFilters] = useState<ProductFilters>({});
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('list');
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [showAdd, setShowAdd] = useState(false);

  const mergedFilters: ProductFilters = { ...filters, search: debouncedSearch };
  const { data: products = [], isLoading } = useProducts(mergedFilters);
  const { data: categories = [] } = useCategories();
  const { data: brands = [] } = useBrands();

  const toggle = (id: number) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelected(next);
  };

  const handleBulkDelete = () => {
    toast.success(`Deleted ${selected.size} products`);
    setSelected(new Set());
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
      className="space-y-5 max-w-7xl"
    >
      <PageHeader
        title="Products"
        description={`${products.length} products in your catalog`}
        actions={
          <>
            <ImportCSV onImport={(rows) => toast.success(`${rows.length} rows queued`)} expectedColumns={['name', 'sku', 'category', 'stock']} />
            <ExportButton data={exportable} filename="products.csv" />
            <button
              onClick={() => setShowAdd(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: AMBER, color: '#000' }}
            >
              <Plus size={14} />
              New Product
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
          {(filters.category || filters.brand || filters.status) && (
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
          title="No products found"
          description="Try adjusting your filters or create your first product."
          action={
            <button
              onClick={() => setShowAdd(true)}
              className="px-3 py-2 rounded-lg text-xs font-semibold"
              style={{ background: AMBER, color: '#000' }}
            >
              Create Product
            </button>
          }
        />
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          {products.map((p) => (
            <ProductCard
              key={p.id}
              product={p}
              selected={selected.has(p.id)}
              onToggle={toggle}
            />
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
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
              >
                <input
                  type="checkbox"
                  checked={selected.has(p.id)}
                  onChange={() => toggle(p.id)}
                  className="cursor-pointer"
                />
                <Link
                  to={`/products/${p.id}`}
                  className="text-sm font-medium hover:underline"
                  style={{ color: 'var(--hub-text)', textDecoration: 'none' }}
                >
                  {p.name}
                </Link>
                <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                  {p.sku}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {p.category}
                </span>
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {p.brand || '—'}
                </span>
                <span
                  className="text-sm font-semibold"
                  style={{
                    color: p.stock === 0 ? '#f87171' : p.stock < p.lowStock ? AMBER : 'var(--hub-text)',
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
          <label className="text-xs font-semibold mb-1 block" style={{ color: 'var(--hub-text-muted)' }}>
            Category
          </label>
          <select
            value={filters.category || ''}
            onChange={(e) => setFilters({ ...filters, category: e.target.value || undefined })}
            className="w-full px-2 py-1.5 rounded-lg text-sm outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          >
            <option value="">All categories</option>
            {categories.map((c) => (
              <option key={c.id} value={c.name}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="text-xs font-semibold mb-1 block" style={{ color: 'var(--hub-text-muted)' }}>
            Brand
          </label>
          <select
            value={filters.brand || ''}
            onChange={(e) => setFilters({ ...filters, brand: e.target.value || undefined })}
            className="w-full px-2 py-1.5 rounded-lg text-sm outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
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
          <label className="text-xs font-semibold mb-1 block" style={{ color: 'var(--hub-text-muted)' }}>
            Status
          </label>
          <select
            value={filters.status || ''}
            onChange={(e) => setFilters({ ...filters, status: (e.target.value as any) || undefined })}
            className="w-full px-2 py-1.5 rounded-lg text-sm outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
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
          toast.success(`Archived ${selected.size} products`);
          setSelected(new Set());
        }}
        onExport={() => {
          toast.success(`Exported ${selected.size} products`);
        }}
      />

      <QuickAddProductModal open={showAdd} onClose={() => setShowAdd(false)} />
    </motion.div>
  );
};

export default Products;
