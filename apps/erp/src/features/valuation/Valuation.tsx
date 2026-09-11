import React, { useMemo } from 'react';
import { motion } from 'motion/react';
import { DollarSign, BarChart3, Boxes } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { useValuation } from './useValuation';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatCurrency, formatNumber } from '../../shared/lib/format';
import { safeRatio, toFiniteNumber } from '@voxire/api';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const AMBER = 'var(--erp-accent)';

const Valuation: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data, isLoading, refetch } = useValuation();
  const items = useMemo(() => data?.items ?? [], [data]);
  const totalCost = useMemo(
    () => items.reduce((sum, item) => sum + (item.totalCost ?? 0), 0),
    [items]
  );
  const totalItems = useMemo(
    () => items.reduce((sum, item) => sum + toFiniteNumber(item.totalItems), 0),
    [items]
  );

  const categories = useMemo(
    () =>
      items
        .map((i) => ({ name: i.categoryName || 'Uncategorized', value: i.totalValue ?? 0 }))
        .sort((a, b) => b.value - a.value),
    [items]
  );
  const maxCat = categories.length > 0 ? categories[0].value : 1;

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/valuation'].title))} />
      <PageHeader
        title={t(routeLabels['/valuation'].title)}
        description="Total value of inventory on hand"
      />

      {isLoading ? (
        <Skeleton className="h-32 rounded-xl" />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <DollarSign size={14} style={{ color: AMBER }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Stock Value
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {formatCurrency(data?.totalValue ?? 0)}
            </p>
          </div>
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <DollarSign size={14} style={{ color: 'var(--erp-success)' }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Cost Basis
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--erp-success)' }}>
              {formatCurrency(totalCost)}
            </p>
          </div>
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 size={14} style={{ color: 'var(--erp-info)' }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Total Items
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {formatNumber(totalItems)}
            </p>
          </div>
        </div>
      )}

      <div
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          Value by Category
        </h3>
        {!isLoading && categories.length === 0 ? (
          <EmptyState
            icon={Boxes}
            title={t('erp.valuation.empty.title')}
            description={t('erp.valuation.empty.description')}
            size="compact"
            action={{ label: t('erp.valuation.refresh'), onClick: () => void refetch() }}
          />
        ) : (
          <div className="space-y-2">
            {categories.map((cat) => (
              <div key={cat.name}>
                <div className="flex justify-between text-xs mb-1">
                  <span style={{ color: 'var(--hub-text-muted)' }}>{cat.name}</span>
                  <span style={{ color: 'var(--hub-text)' }}>{formatCurrency(cat.value)}</span>
                </div>
                <div className="h-1.5 rounded-full" style={{ background: 'var(--hub-hover)' }}>
                  <div
                    className="h-full rounded-full"
                    style={{
                      width: `${safeRatio(cat.value, maxCat) * 100}%`,
                      background: AMBER,
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </motion.div>
  );
};

export default Valuation;
