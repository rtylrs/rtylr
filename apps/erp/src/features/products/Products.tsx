import React, { useCallback, useMemo, useState } from 'react';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { cdnImageUrl } from '@voxire/config';
import { Search, Plus, Grid3X3, List, SlidersHorizontal, ArrowUpDown, Package } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useDebounce } from '@voxire/utils';
import {
  getProductStockStatus,
  useProducts,
  useDeleteProduct,
  type Product,
  type ProductFilters,
} from './useProducts';
import { useCategories } from '../categories/useCategories';
import ProductCard from '../../shared/components/ProductCard';
import StockBadge from '../../shared/components/StockBadge';
import FilterDrawer from '../../shared/components/FilterDrawer';
import BulkActionBar from '../../shared/components/BulkActionBar';
import ExportButton from '../../shared/components/ExportButton';
import ImportCSV from '../../shared/components/ImportCSV';
import EmptyState from '../../shared/components/EmptyState';
import PaginationControls from '../../shared/components/PaginationControls';
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
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showAdd, setShowAdd] = useState(false);
  const [page, setPage] = useState(1);
  const pageSize = 50;

  const mergedFilters: ProductFilters = { ...filters, search: debouncedSearch, page, pageSize };
  const { data: productResult, isLoading } = useProducts(mergedFilters);
  const rawProducts = productResult?.products ?? [];
  const products = useMemo(() => {
    let out = rawProducts;

    if (filters.status) {
      out = out.filter((product) => getProductStockStatus(product) === filters.status);
    }

    if (filters.sort) {
      const direction = filters.sortDir === 'desc' ? -1 : 1;
      out = [...out].sort((a, b) => {
        if (filters.sort === 'name') return a.name.localeCompare(b.name) * direction;
        if (filters.sort === 'stock') return (a.stockQuantity - b.stockQuantity) * direction;
        if (filters.sort === 'price') return (a.salePrice - b.salePrice) * direction;
        if (filters.sort === 'category') return a.categoryId.localeCompare(b.categoryId) * direction;
        return 0;
      });
    }

    return out;
  }, [rawProducts, filters.status, filters.sort, filters.sortDir]);
  const totalProducts = productResult?.total ?? products.length;
  const hasMore = productResult?.hasMore ?? false;
  const { data: categories = [] } = useCategories();
  const deleteMut = useDeleteProduct();
  const categoryNameById = useMemo(() => {
    const map = new Map<string, string>();
    categories.forEach((category) => map.set(String(category.id), category.name));
    return map;
  }, [categories]);

  const categoryNameFor = useCallback(
    (product: Product): string => {
      return categoryNameById.get(product.categoryId) || '-';
    },
    [categoryNameById]
  );

  const changePage = (nextPage: number): void => {
    setSelected(new Set());
    setPage(nextPage);
  };

  const updateFilters = (next: ProductFilters): void => {
    setSelected(new Set());
    setPage(1);
    setFilters(next);
  };

  const toggle = (id: string) => {
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
        category: categoryNameFor(p),
        brand: '',
        stock: p.stockQuantity,
        costPrice: p.costPrice,
        salePrice: p.salePrice,
        status: getProductStockStatus(p),
      })),
    [products, categoryNameFor]
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
        description={`${totalProducts} product${totalProducts === 1 ? '' : 's'} in your catalog`}
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
            onChange={(e) => {
              setPage(1);
              setSelected(new Set());
              setSearch(e.target.value);
            }}
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
          {(filters.categoryId !== undefined || filters.status) && (
            <span className="h-1.5 w-1.5 rounded-full" style={{ background: AMBER }} />
          )}
        </button>
        <button
          onClick={() =>
            updateFilters({
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
            {products.map((p) => {
              const imageUrl = cdnImageUrl((p as Product & { imageUrl?: string }).imageUrl);
              const stockStatus = getProductStockStatus(p);
              return (
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
                      {imageUrl ? (
                        <img
                          src={imageUrl}
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
                  <span className="truncate text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {categoryNameFor(p)}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    -
                  </span>
                  <span
                    className="text-sm font-semibold"
                    style={{
                      color:
                        p.stockQuantity === 0
                          ? '#f87171'
                          : p.stockQuantity < p.lowStockThreshold
                            ? AMBER
                            : 'var(--hub-text)',
                    }}
                  >
                    {p.stockQuantity}
                  </span>
                  <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                    {formatCurrency(p.salePrice)}
                  </span>
                  <StockBadge stock={p.stockQuantity} lowStock={p.lowStockThreshold} />
                </div>
              );
            })}
          </div>
        </div>
      )}

      <PaginationControls
        page={page}
        hasMore={hasMore}
        itemCount={products.length}
        pageSize={pageSize}
        total={totalProducts}
        onPageChange={changePage}
      />

      <FilterDrawer
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        onReset={() => updateFilters({})}
      >
        <div>
          <label
            className="text-xs font-semibold mb-1 block"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            Category
          </label>
          <select
            value={filters.categoryId ?? ''}
            onChange={(e) =>
              updateFilters({
                ...filters,
                categoryId: e.target.value || undefined,
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
            Status
          </label>
          <select
            value={filters.status || ''}
            onChange={(e) =>
              updateFilters({
                ...filters,
                status: (e.target.value as ProductFilters['status']) || undefined,
              })
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
              categoryNameFor(p),
              '',
              p.stockQuantity,
              p.costPrice,
              p.salePrice,
              getProductStockStatus(p),
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
