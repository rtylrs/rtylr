import React from 'react';
import { motion } from 'motion/react';
import { Package, TrendingUp, TrendingDown, BarChart3 } from 'lucide-react';
import { useProductPerformance, useBestSellers, useSlowMovers, useMarginAnalysis, useCategoryBreakdown } from './useProducts';
import { formatCurrency, formatNumber, formatPercent } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const ACCENT = '#6366f1';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Products: React.FC = () => {
  const { data: performance, isLoading: perfLoading, isError, refetch } = useProductPerformance();
  const { data: bestSellers, isLoading: bestLoading } = useBestSellers(10);
  const { data: slowMovers, isLoading: slowLoading } = useSlowMovers(10);
  const { data: margins, isLoading: marginsLoading } = useMarginAnalysis();
  const { data: categories, isLoading: catLoading } = useCategoryBreakdown();

  if (isError) {
    return <ErrorState message="Failed to load product data." onRetry={refetch} />;
  }

  const bestList: any[] = bestSellers?.data ?? bestSellers ?? [];
  const slowList: any[] = slowMovers?.data ?? slowMovers ?? [];
  const marginData = margins ?? {};
  const catList: any[] = categories?.data ?? categories ?? [];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>Products</h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          Product performance and margin analysis
        </p>
      </motion.div>

      {/* Summary cards */}
      {perfLoading || marginsLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Total Products', value: formatNumber(performance?.totalProducts ?? 0), icon: Package, color: ACCENT },
            { label: 'Avg Margin', value: `${(marginData.averageMargin ?? 0).toFixed(1)}%`, icon: BarChart3, color: '#10b981' },
            { label: 'Best Sellers', value: formatNumber(bestList.length), icon: TrendingUp, color: '#f59e0b' },
            { label: 'Slow Movers', value: formatNumber(slowList.length), icon: TrendingDown, color: '#f87171' },
          ].map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>{stat.label}</span>
                  <span className="flex items-center justify-center h-7 w-7 rounded-lg" style={{ background: `${stat.color}18` }}>
                    <Icon size={14} style={{ color: stat.color }} />
                  </span>
                </div>
                <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>{stat.value}</p>
              </motion.div>
            );
          })}
        </motion.div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Best Sellers */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp size={14} style={{ color: '#10b981' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Best Sellers</h3>
          </div>
          {bestLoading ? (
            <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={3} />)}</div>
          ) : bestList.length === 0 ? (
            <EmptyState title="No data" description="Best sellers will appear here." />
          ) : (
            <div className="space-y-1">
              {bestList.map((p: any, i: number) => (
                <div key={p.id ?? i} className="flex items-center justify-between px-3 py-2.5 rounded-lg" style={{ background: 'var(--hub-hover)' }}>
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] w-5" style={{ color: 'var(--hub-text-dim)' }}>#{i + 1}</span>
                    <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>{p.name ?? 'Unknown'}</span>
                  </div>
                  <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>{formatCurrency(p.revenue ?? p.sales ?? 0)}</span>
                </div>
              ))}
            </div>
          )}
        </motion.div>

        {/* Slow Movers */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-4">
            <TrendingDown size={14} style={{ color: '#f87171' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Slow Movers</h3>
          </div>
          {slowLoading ? (
            <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={3} />)}</div>
          ) : slowList.length === 0 ? (
            <EmptyState title="No data" description="Slow movers will appear here." />
          ) : (
            <div className="space-y-1">
              {slowList.map((p: any, i: number) => (
                <div key={p.id ?? i} className="flex items-center justify-between px-3 py-2.5 rounded-lg" style={{ background: 'var(--hub-hover)' }}>
                  <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>{p.name ?? 'Unknown'}</span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{formatNumber(p.unitsSold ?? p.quantity ?? 0)} sold</span>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      </div>

      {/* Category Breakdown */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>Category Breakdown</h3>
        {catLoading ? (
          <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={3} />)}</div>
        ) : catList.length === 0 ? (
          <EmptyState title="No categories" description="Category breakdown will appear here." />
        ) : (
          <div className="space-y-1">
            {catList.map((c: any, i: number) => (
              <div key={c.id ?? i} className="flex items-center justify-between px-3 py-2.5 rounded-lg" style={{ background: 'var(--hub-hover)' }}>
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{c.name ?? c.category ?? 'Uncategorized'}</span>
                <div className="flex items-center gap-4">
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{formatNumber(c.productCount ?? c.count ?? 0)} products</span>
                  <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>{formatCurrency(c.revenue ?? 0)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Products;
