import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { Search, Plus } from 'lucide-react';
import { erpApi, Product } from '../../services/erpApi';

const AMBER = '#f59e0b';

// ─── Mock / fallback data ─────────────────────────────────────────────────────

const mockProducts: Product[] = [
  { id: 1, name: 'Arabica Coffee Beans',       sku: 'BEV-001',  category: 'Beverages', stock: 3,   lowStock: 20, costPrice: 18.5, salePrice: 29.99, status: 'Low Stock'    },
  { id: 2, name: 'Whole Milk (2L)',             sku: 'BEV-012',  category: 'Beverages', stock: 8,   lowStock: 50, costPrice: 2.1,  salePrice: 3.5,   status: 'Low Stock'    },
  { id: 3, name: 'Sourdough Bread',             sku: 'FOOD-034', category: 'Bakery',    stock: 0,   lowStock: 15, costPrice: 3.0,  salePrice: 5.5,   status: 'Out of Stock' },
  { id: 4, name: 'Extra Virgin Olive Oil (1L)', sku: 'FOOD-055', category: 'Pantry',    stock: 42,  lowStock: 10, costPrice: 7.8,  salePrice: 14.99, status: 'In Stock'     },
  { id: 5, name: 'Sparkling Water (500ml)',     sku: 'BEV-028',  category: 'Beverages', stock: 120, lowStock: 30, costPrice: 0.6,  salePrice: 1.5,   status: 'In Stock'     },
  { id: 6, name: 'Cheddar Cheese (500g)',       sku: 'DAIRY-007',category: 'Dairy',     stock: 18,  lowStock: 15, costPrice: 5.2,  salePrice: 9.99,  status: 'In Stock'     },
];

// ─── Config ───────────────────────────────────────────────────────────────────

const statusConfig: Record<Product['status'], { color: string; bg: string }> = {
  'In Stock':      { color: '#10b981', bg: 'rgba(16,185,129,0.1)'  },
  'Low Stock':     { color: AMBER,     bg: 'rgba(245,158,11,0.1)'  },
  'Out of Stock':  { color: '#f87171', bg: 'rgba(248,113,113,0.1)' },
};

// ─── Animation variants ───────────────────────────────────────────────────────

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const columns = ['Name', 'SKU', 'Category', 'Stock', 'Low Stock', 'Cost Price', 'Sale Price', 'Status'];

// ─── Skeleton row ─────────────────────────────────────────────────────────────

const SkeletonTableRow: React.FC = () => (
  <div
    className="grid px-4 py-3 items-center"
    style={{ gridTemplateColumns: '2fr 1fr 1fr 0.7fr 0.7fr 1fr 1fr 0.9fr' }}
  >
    {Array.from({ length: 8 }).map((_, i) => (
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
  const [search, setSearch] = useState('');
  const [products, setProducts] = useState<Product[]>(mockProducts);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function fetchProducts() {
      setIsLoading(true);
      try {
        const res = await erpApi.listProducts({ search: search || undefined });
        if (!cancelled && res.data?.length) {
          setProducts(res.data);
        }
      } catch {
        // silently fall back to mock data
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    }

    fetchProducts();
    return () => { cancelled = true; };
  }, []);

  // Client-side search filters mock data; real API uses search param on mount
  const filtered = isLoading
    ? []
    : products.filter(
        (p) =>
          p.name.toLowerCase().includes(search.toLowerCase()) ||
          p.sku.toLowerCase().includes(search.toLowerCase()) ||
          p.category.toLowerCase().includes(search.toLowerCase())
      );

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-5 max-w-6xl"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Inventory
        </h2>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          style={{ background: AMBER, color: '#000' }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.opacity = '0.9'; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.opacity = '1'; }}
        >
          <Plus size={14} />
          Add Product
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
          placeholder="Search products…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-9 pr-4 py-2 rounded-lg text-sm outline-none transition-colors"
          style={{
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text)',
          }}
          onFocus={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'rgba(245,158,11,0.5)'; }}
          onBlur={(e)  => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)'; }}
        />
      </motion.div>

      {/* Table */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl overflow-hidden"
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
        }}
      >
        {/* Column headers */}
        <div
          className="grid px-4 py-3"
          style={{
            gridTemplateColumns: '2fr 1fr 1fr 0.7fr 0.7fr 1fr 1fr 0.9fr',
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
          ) : filtered.length === 0 ? (
            <div className="px-4 py-8 text-center text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              No products found.
            </div>
          ) : (
            filtered.map((product) => {
              const sc = statusConfig[product.status];
              return (
                <div
                  key={product.id}
                  className="grid px-4 py-3 items-center transition-colors"
                  style={{ gridTemplateColumns: '2fr 1fr 1fr 0.7fr 0.7fr 1fr 1fr 0.9fr' }}
                  onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                >
                  <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                    {product.name}
                  </span>
                  <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                    {product.sku}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {product.category}
                  </span>
                  <span
                    className="text-sm font-semibold"
                    style={{
                      color:
                        product.stock === 0
                          ? '#f87171'
                          : product.stock < product.lowStock
                          ? AMBER
                          : 'var(--hub-text)',
                    }}
                  >
                    {product.stock}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {product.lowStock}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    ${product.costPrice.toFixed(2)}
                  </span>
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    ${product.salePrice.toFixed(2)}
                  </span>
                  <span>
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{ background: sc.bg, color: sc.color }}
                    >
                      {product.status}
                    </span>
                  </span>
                </div>
              );
            })
          )}
        </div>
      </motion.div>
    </motion.div>
  );
};

export default Inventory;
