import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { useCartStore } from '../../../../stores/cartStore';
import styles from './ProductGrid.module.scss';

interface Product {
  id: string;
  name: string;
  price: number;
  category: string;
  image?: string;
  available?: boolean;
}

interface ProductGridProps {
  products: Product[];
  categories: string[];
}

const ProductGrid: React.FC<ProductGridProps> = ({ products, categories }) => {
  const [activeCategory, setActiveCategory] = useState<string>('All');
  const [search, setSearch] = useState('');
  const addItem = useCartStore((s) => s.addItem);

  const allCategories = ['All', ...categories];

  const filtered = products.filter((p) => {
    const matchCat = activeCategory === 'All' || p.category === activeCategory;
    const matchSearch = p.name.toLowerCase().includes(search.toLowerCase());
    return matchCat && matchSearch;
  });

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);

  return (
    <div className={styles.grid}>
      {/* Search + Category Bar */}
      <div className={styles.grid__toolbar}>
        <input
          className={styles.grid__search}
          type="search"
          placeholder="Search items…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className={styles.grid__categories}>
          {allCategories.map((cat) => (
            <button
              key={cat}
              className={`${styles.grid__catBtn} ${activeCategory === cat ? styles['grid__catBtn--active'] : ''}`}
              onClick={() => setActiveCategory(cat)}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {/* Products */}
      <div className={styles.grid__items}>
        {filtered.length === 0 ? (
          <div className={styles.grid__empty}>No items found</div>
        ) : (
          filtered.map((product) => (
            <motion.button
              key={product.id}
              className={`${styles.grid__item} ${product.available === false ? styles['grid__item--unavailable'] : ''}`}
              onClick={() => product.available !== false && addItem(product)}
              whileTap={{ scale: 0.96 }}
              disabled={product.available === false}
            >
              {product.image && (
                <div className={styles.grid__itemImg}>
                  <img src={product.image} alt={product.name} loading="lazy" />
                </div>
              )}
              <div className={styles.grid__itemBody}>
                <span className={styles.grid__itemName}>{product.name}</span>
                <span className={styles.grid__itemPrice}>{formatPrice(product.price)}</span>
              </div>
              {product.available === false && (
                <span className={styles.grid__itemTag}>Out of stock</span>
              )}
            </motion.button>
          ))
        )}
      </div>
    </div>
  );
};

export default ProductGrid;
