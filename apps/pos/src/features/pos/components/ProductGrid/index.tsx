import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useCartStore } from '../../../../shared/stores/cartStore';
import { useSettingsStore } from '../../../../shared/stores/settingsStore';
import { buildMenuItemCartProduct } from '../../../../shared/utils/cart/cartProduct';
import { formatCurrency } from '../../../../shared/utils/format';
import type { MenuItem } from '../../../../shared/types/pos';

type Product = MenuItem;
interface ProductGridProps {
  products: Product[];
  categories: string[];
}

const hasFiniteStock = (stock: unknown): stock is number =>
  typeof stock === 'number' && Number.isFinite(stock);

const formatStock = (stock: number) =>
  Number.isInteger(stock) ? String(stock) : stock.toFixed(3).replace(/\.?0+$/, '');

const ProductGrid: React.FC<ProductGridProps> = ({ products, categories }) => {
  const [activeCategory, setActiveCategory] = useState('All');
  const [search, setSearch] = useState('');
  const addItem = useCartStore((s) => s.addItem);
  const businessProfile = useSettingsStore((s) => s.businessProfile);
  const inventoryEnabled = useSettingsStore((s) => s.features.inventory);
  const currency = useSettingsStore((s) => s.currency);

  const allCategories = ['All', ...categories];
  const itemNoun = businessProfile === 'retail' ? 'products' : 'items';
  const filtered = products.filter((p) => {
    const matchCat = activeCategory === 'All' || p.category === activeCategory;
    const searchTerm = search.trim().toLowerCase();
    const matchSearch =
      !searchTerm ||
      [p.name, p.sku, p.barcode, p.category]
        .filter((value): value is string => Boolean(value))
        .some((value) => value.toLowerCase().includes(searchTerm));
    return matchCat && matchSearch;
  });

  const addProduct = (product: Product) => {
    if (product.available === false) return;

    try {
      addItem(buildMenuItemCartProduct(product));
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : `Failed to add ${itemNoun.slice(0, -1)}`
      );
    }
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Toolbar */}
      <div className="p-3 border-b border-pos-border space-y-2 shrink-0">
        <input
          className="w-full rounded-lg bg-white/6 border border-pos-border px-3.5 py-2 text-sm text-pos-text placeholder:text-pos-muted outline-none focus:border-white/20 transition-colors"
          type="search"
          placeholder={`Search ${itemNoun}...`}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="flex gap-1.5 overflow-x-auto pb-0.5 scrollbar-none">
          {allCategories.map((cat) => (
            <button
              key={cat}
              className={`shrink-0 px-3 py-1.5 rounded-full text-xs font-semibold transition-colors ${
                activeCategory === cat
                  ? 'bg-pos-text text-pos-bg'
                  : 'bg-white/6 text-pos-muted hover:bg-white/10'
              }`}
              onClick={() => setActiveCategory(cat)}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Grid */}
      <div className="flex-1 overflow-y-auto p-3">
        {filtered.length === 0 ? (
          <div className="flex items-center justify-center h-40 text-pos-muted text-sm">
            No {itemNoun} found
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-2.5">
            {filtered.map((product) => {
              const unavailable = product.available === false;
              const stock = product.stock;
              const showStock = inventoryEnabled && hasFiniteStock(stock);
              return (
                <motion.button
                  key={product.id}
                  className={`flex min-h-[13rem] flex-col overflow-hidden rounded-xl bg-pos-card border border-pos-border text-left transition-colors hover:border-white/15 ${
                    unavailable ? 'opacity-50 cursor-not-allowed' : ''
                  }`}
                  onClick={() => addProduct(product)}
                  whileTap={{ scale: 0.96 }}
                  disabled={unavailable}
                >
                  {product.image ? (
                    <img
                      src={product.image}
                      alt={product.name}
                      className="w-full aspect-[4/3] object-contain bg-white p-2"
                      loading="lazy"
                    />
                  ) : (
                    <div className="flex w-full aspect-[4/3] items-center justify-center bg-white/6 text-lg font-semibold text-pos-muted">
                      {product.name.slice(0, 1).toUpperCase()}
                    </div>
                  )}
                  <div className="flex flex-1 flex-col gap-1.5 p-3">
                    <span className="text-sm font-semibold text-pos-text leading-tight line-clamp-2">
                      {product.name}
                    </span>
                    {product.sku && (
                      <span className="truncate text-[10px] font-medium uppercase tracking-wide text-pos-muted/80">
                        {product.sku}
                      </span>
                    )}
                    <div className="mt-auto flex items-end justify-between gap-2">
                      <span className="text-sm font-semibold text-pos-text">
                        {formatCurrency(product.price, currency)}
                      </span>
                      {showStock && (
                        <span className="rounded-full bg-white/6 px-2 py-1 text-[10px] font-semibold text-pos-muted">
                          {formatStock(stock)} in stock
                        </span>
                      )}
                    </div>
                    {unavailable && (
                      <span className="text-[10px] text-destructive font-semibold">
                        Unavailable
                      </span>
                    )}
                  </div>
                </motion.button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProductGrid;
