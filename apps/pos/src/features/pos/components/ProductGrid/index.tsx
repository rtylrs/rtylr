import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useCartStore } from '../../../../shared/stores/cartStore';
import { buildMenuItemCartProduct } from '../../../../shared/utils/cart/cartProduct';
import type { MenuItem } from '../../../../shared/types/pos';

type Product = MenuItem;
interface ProductGridProps {
  products: Product[];
  categories: string[];
}

const fmt = (n: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const ProductGrid: React.FC<ProductGridProps> = ({ products, categories }) => {
  const [activeCategory, setActiveCategory] = useState('All');
  const [search, setSearch] = useState('');
  const addItem = useCartStore((s) => s.addItem);

  const allCategories = ['All', ...categories];
  const filtered = products.filter((p) => {
    const matchCat = activeCategory === 'All' || p.category === activeCategory;
    const matchSearch = p.name.toLowerCase().includes(search.toLowerCase());
    return matchCat && matchSearch;
  });

  const addProduct = (product: Product) => {
    if (product.available === false) return;

    try {
      addItem(buildMenuItemCartProduct(product));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to add menu item');
    }
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Toolbar */}
      <div className="p-3 border-b border-pos-border space-y-2 shrink-0">
        <input
          className="w-full rounded-lg bg-white/6 border border-pos-border px-3.5 py-2 text-sm text-pos-text placeholder:text-pos-muted outline-none focus:border-white/20 transition-colors"
          type="search"
          placeholder="Search items…"
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
            No items found
          </div>
        ) : (
          <div className="grid grid-cols-3 gap-2.5">
            {filtered.map((product) => (
              <motion.button
                key={product.id}
                className={`flex flex-col p-3 rounded-xl bg-pos-card border border-pos-border text-left transition-colors hover:border-white/15 ${
                  product.available === false ? 'opacity-50 cursor-not-allowed' : ''
                }`}
                onClick={() => addProduct(product)}
                whileTap={{ scale: 0.96 }}
                disabled={product.available === false}
              >
                {product.image && (
                  <img
                    src={product.image}
                    alt={product.name}
                    className="w-full aspect-square object-cover rounded-lg mb-2"
                    loading="lazy"
                  />
                )}
                <span className="text-sm font-medium text-pos-text leading-tight">
                  {product.name}
                </span>
                <span className="text-xs text-pos-muted mt-1">{fmt(product.price)}</span>
                {product.available === false && (
                  <span className="mt-1 text-[10px] text-destructive font-semibold">
                    Out of stock
                  </span>
                )}
              </motion.button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProductGrid;
