import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion, type Variants } from 'motion/react';
import { Package, TrendingDown, BarChart3, Layers } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import {
  useProductKPIs,
  useBestSellers,
  useSlowMovers,
  useProductMargins,
  useProductCategoryMix,
} from '../../shared/hooks/queries/useInsightsProducts';
import { useInsightsSettings } from '../../shared/hooks/queries/useInsightsSettings';
import { formatMoney, type Currency } from '../../shared/utils/money';
import { formatNumber } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Products: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: kpis, isLoading: kpisLoading, isError, refetch } = useProductKPIs();
  const {
    data: bestRes,
    isLoading: bestLoading,
    refetch: refetchBest,
  } = useBestSellers({ limit: 10 });
  const {
    data: slowRes,
    isLoading: slowLoading,
    refetch: refetchSlow,
  } = useSlowMovers({ days: 60 });
  const {
    data: marginRes,
    isLoading: marginLoading,
    refetch: refetchMargins,
  } = useProductMargins();
  const {
    data: catRes,
    isLoading: catLoading,
    refetch: refetchCategories,
  } = useProductCategoryMix();
  const { data: settings } = useInsightsSettings();
  const currency = (settings?.defaultCurrency as Currency) || 'USD';

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
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Products · rtylr" />
      <motion.div variants={itemVariants}>
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.title')}
        </h1>
      </motion.div>

      {kpisLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants}
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
                    <Icon size={14} style={{ color: 'var(--accent-on-muted)' }} />
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
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.bestSellers')}
        </h2>
        {bestLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : best.length === 0 ? (
          <EmptyState
            icon={Package}
            title={t('insights.products.emptyBestSellers.title')}
            description={t('insights.products.emptyBestSellers.description')}
            action={{ label: t('common.refresh'), onClick: () => refetchBest() }}
          />
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
                  {t('insights.products.units', {
                    count: p.unitsSold,
                    formatted: formatNumber(p.unitsSold),
                  })}
                </span>
                <span
                  className="text-xs font-semibold text-right"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {formatMoney(p.revenueMinorUnits, currency)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.margins')}
        </h2>
        {marginLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : margins.length === 0 ? (
          <EmptyState
            icon={BarChart3}
            title={t('insights.products.emptyMargins.title')}
            description={t('insights.products.emptyMargins.description')}
            action={{ label: t('common.refresh'), onClick: () => refetchMargins() }}
          />
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
                  {t('insights.products.products', {
                    count: b.productCount,
                    formatted: formatNumber(b.productCount),
                  })}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.slowMovers')}
        </h2>
        {slowLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : slow.length === 0 ? (
          <EmptyState
            icon={TrendingDown}
            title={t('insights.products.emptySlowMovers.title')}
            description={t('insights.products.emptySlowMovers.description')}
            action={{ label: t('common.refresh'), onClick: () => refetchSlow() }}
          />
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
                  {formatNumber(p.stockQuantity)} {t('insights.products.inStock')}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.products.categoryMix')}
        </h2>
        {catLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : categories.length === 0 ? (
          <EmptyState
            icon={Layers}
            title={t('insights.products.emptyCategoryMix.title')}
            description={t('insights.products.emptyCategoryMix.description')}
            action={{ label: t('common.refresh'), onClick: () => refetchCategories() }}
          />
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
                  {t('insights.products.units', {
                    count: c.unitsSold,
                    formatted: formatNumber(c.unitsSold),
                  })}
                </span>
                <span
                  className="text-xs font-semibold text-right"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {formatMoney(c.revenueMinorUnits, currency)}
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
