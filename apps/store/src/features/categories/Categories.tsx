import React from 'react';
import { motion } from 'framer-motion';
import { Folder, Plus, Package } from 'lucide-react';
import { useCategories } from './useCategories';
import { SkeletonRow } from '../../shared/components/Skeleton';

const TEAL = '#14b8a6';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.03 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.18, ease: 'easeOut' } },
};

const Categories: React.FC = () => {
  const { data: categories, isLoading, error } = useCategories();

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>Failed to load categories</p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>Please check your connection and try again.</p>
      </div>
    );
  }

  return (
    <motion.div variants={containerVariants} initial="hidden" animate="show" className="space-y-4 max-w-5xl">
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Folder size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Categories</h2>
          {!isLoading && (
            <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}>
              {(categories ?? []).length}
            </span>
          )}
        </div>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: TEAL, color: '#fff' }}
        >
          <Plus size={14} />
          Add Category
        </button>
      </motion.div>

      {isLoading ? (
        <div className="space-y-1">
          {Array.from({ length: 6 }).map((_, i) => <SkeletonRow key={i} cols={4} />)}
        </div>
      ) : (categories ?? []).length === 0 ? (
        <motion.div variants={itemVariants} className="flex flex-col items-center py-16 gap-3">
          <Folder size={32} style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>No categories yet</p>
          <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>Create your first product category to organize your catalog.</p>
        </motion.div>
      ) : (
        <motion.div variants={containerVariants} className="space-y-1">
          <div className="grid grid-cols-4 gap-2 px-3 py-2 text-[10px] uppercase tracking-wider font-semibold"
            style={{ color: 'var(--hub-text-dim)' }}>
            <span>Category</span>
            <span>Slug</span>
            <span>Products</span>
            <span>Sort Order</span>
          </div>
          {(categories ?? []).map((cat) => (
            <motion.div
              key={cat.id}
              variants={itemVariants}
              className="grid grid-cols-4 gap-2 items-center px-3 py-2.5 rounded-lg cursor-pointer"
              style={{ background: 'var(--hub-hover)' }}
            >
              <div className="flex items-center gap-3 min-w-0">
                <div className="h-7 w-7 rounded-lg shrink-0 flex items-center justify-center"
                  style={{ background: 'rgba(20,184,166,0.08)' }}>
                  <Folder size={12} style={{ color: TEAL }} />
                </div>
                <div className="min-w-0">
                  <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>{cat.name}</p>
                  {cat.description && (
                    <p className="text-[10px] truncate" style={{ color: 'var(--hub-text-dim)' }}>{cat.description}</p>
                  )}
                </div>
              </div>
              <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{cat.slug}</span>
              <div className="flex items-center gap-1">
                <Package size={10} style={{ color: 'var(--hub-text-dim)' }} />
                <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{cat.productCount}</span>
              </div>
              <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{cat.sortOrder}</span>
            </motion.div>
          ))}
        </motion.div>
      )}
    </motion.div>
  );
};

export default Categories;
