import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Package, TrendingDown, BarChart3 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useProductKPIs,
  useBestSellers,
  useSlowMovers,
  useProductMargins,
  useProductCategoryMix,
} from '../../shared/hooks/queries/useInsightsProducts';
import { formatMoney } from '../../shared/utils/money';
import { formatNumber } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Products: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: kpis, isLoading: kpisLoading, isError, refetch } = useProductKPIs();
  const { data: bestRes, isLoading: bestLoading } = useBestSellers({ limit: 10 });
  const { data: slowRes, isLoading: slowLoading } = useSlowMovers({ days: 60 });
  const { data: marginRes, isLoading: marginLoading } = useProductMargins();
  const { data: catRes, isLoading: catLoading } = useProductCategoryMix();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const totalSkus = kpis?.totalSkus ?? 0;
  const activeSkus = kpis?.activeSkus ?? 0;
  const outOfStock = kpis?.outOfStock ?? 0;
  const avgMarginPct = kpis?.avgMarginPct ?? 0;
  const best = bestRes?.products ?? [];
  const slow = slowRes?.products ?? [];
  const margins = marginRes?.buckets ?? [];
  const categories = catRes?.categories ?? [];

  const stats = [
    {
      label: t('insights.products.kpis.totalSkus'),
      value: formatNumber(totalSkus),
      icon: Package,
    },
    {
      label: t('insights.products.kpis.activeSkus'),
      value: formatNumber(activeSkus),
      icon: BarChart3,
    },
    {
      label: t('insights.products.kpis.outOfStock'),
      value: formatNumber(outOfStock),
      icon: TrendingDown,
    },
    {
      label: t('insights.products.kpis.avgMargin'),
      value: `${avgMarginPct.toFixed(1)}%`,
      icon: BarChart3,
    },
  ];

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Products · rtylr" />
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.title')}
        </h2>
      </motion.div>

      {kpisLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div
          variants={containerVariants as any}
          className="grid grid-cols-2 lg:grid-cols-4 gap-4"
        >
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants as any}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                    {stat.label}
                  </span>
                  <span
                    className="flex items-center justify-center h-7 w-7 rounded-lg"
                    style={{ background: 'var(--accent-muted)' }}
                  >
                    <Icon size={14} style={{ color: 'var(--accent)' }} />
                  </span>
                </div>
                <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                  {stat.value}
                </p>
              </motion.div>
            );
          })}
        </motion.div>
      )}

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.bestSellers')}
        </h3>
        {bestLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : best.length === 0 ? (
          <EmptyState title="" description={t('insights.products.empty')} />
        ) : (
          <div className="space-y-1">
            {best.map((p, i) => (
              <div
                key={p.productId}
                className="grid grid-cols-3 items-center px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div className="flex items-center gap-2">
                  <span className="text-[10px] w-5" style={{ color: 'var(--hub-text-dim)' }}>
                    #{i + 1}
                  </span>
                  <span
                    className="text-xs font-medium truncate"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {p.name}
                  </span>
                </div>
                <span className="text-xs text-center" style={{ color: 'var(--hub-text-muted)' }}>
                  {formatNumber(p.unitsSold)} units
                </span>
                <span
                  className="text-xs font-semibold text-right"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {formatMoney(p.revenueMinorUnits)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.margins')}
        </h3>
        {marginLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : margins.length === 0 ? (
          <EmptyState title="" description={t('insights.products.empty')} />
        ) : (
          <div className="space-y-1">
            {margins.map((b) => (
              <div
                key={b.label}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {b.label}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatNumber(b.productCount)} products
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.slowMovers')}
        </h3>
        {slowLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : slow.length === 0 ? (
          <EmptyState title="" description={t('insights.products.empty')} />
        ) : (
          <div className="space-y-1">
            {slow.map((p) => (
              <div
                key={p.productId}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                  {p.name}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatNumber(p.stockQuantity)} in stock
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.categoryMix')}
        </h3>
        {catLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : categories.length === 0 ? (
          <EmptyState title="" description={t('insights.products.empty')} />
        ) : (
          <div className="space-y-1">
            {categories.map((c) => (
              <div
                key={c.categoryId}
                className="grid grid-cols-3 items-center px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                  {c.categoryName}
                </span>
                <span className="text-xs text-center" style={{ color: 'var(--hub-text-muted)' }}>
                  {formatNumber(c.unitsSold)} units
                </span>
                <span
                  className="text-xs font-semibold text-right"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {formatMoney(c.revenueMinorUnits)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Products;
