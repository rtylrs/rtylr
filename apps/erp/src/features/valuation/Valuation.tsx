import React, { useMemo, useState } from 'react';
import { motion } from 'motion/react';
import { DollarSign, BarChart3 } from 'lucide-react';
import { useProducts } from '../products/useProducts';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatCurrency, formatNumber } from '../../shared/lib/format';

const AMBER = '#f59e0b';

type Method = 'FIFO' | 'LIFO' | 'WAC';

const Valuation: React.FC = () => {
  const { data: productResult, isLoading } = useProducts();
  const products = productResult?.products ?? [];
  const [method, setMethod] = useState<Method>('WAC');

  const summary = useMemo(() => {
    const totalValue = products.reduce((s, p) => s + p.stock * p.costPrice, 0);
    const totalRetail = products.reduce((s, p) => s + p.stock * p.salePrice, 0);
    const totalUnits = products.reduce((s, p) => s + p.stock, 0);
    const byCategory = products.reduce<Record<string, number>>((acc, p) => {
      acc[p.category] = (acc[p.category] || 0) + p.stock * p.costPrice;
      return acc;
    }, {});
    return { totalValue, totalRetail, totalUnits, byCategory };
  }, [products]);

  const categories = Object.entries(summary.byCategory).sort((a, b) => b[1] - a[1]);
  const maxCat = categories.length > 0 ? categories[0][1] : 1;

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-6xl"
    >
      <PageHeader
        title="Inventory Valuation"
        description="Total value of inventory on hand"
        actions={
          <div
            className="flex rounded-lg p-0.5"
            style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
          >
            {(['FIFO', 'LIFO', 'WAC'] as Method[]).map((m) => (
              <button
                key={m}
                onClick={() => setMethod(m)}
                className="px-3 py-1 rounded text-xs font-semibold"
                style={{
                  background: method === m ? 'var(--hub-active)' : 'transparent',
                  color: method === m ? AMBER : 'var(--hub-text-muted)',
                }}
              >
                {m}
              </button>
            ))}
          </div>
        }
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
                Cost Value ({method})
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {formatCurrency(summary.totalValue)}
            </p>
          </div>
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <DollarSign size={14} style={{ color: '#10b981' }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Retail Value
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: '#10b981' }}>
              {formatCurrency(summary.totalRetail)}
            </p>
            <p className="text-[11px] mt-1" style={{ color: 'var(--hub-text-dim)' }}>
              Potential margin{' '}
              {Math.round(((summary.totalRetail - summary.totalValue) / summary.totalRetail) * 100)}
              %
            </p>
          </div>
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 size={14} style={{ color: '#60a5fa' }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Total Units
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {formatNumber(summary.totalUnits)}
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
        <div className="space-y-2">
          {categories.map(([cat, value]) => (
            <div key={cat}>
              <div className="flex justify-between text-xs mb-1">
                <span style={{ color: 'var(--hub-text-muted)' }}>{cat}</span>
                <span style={{ color: 'var(--hub-text)' }}>{formatCurrency(value)}</span>
              </div>
              <div className="h-1.5 rounded-full" style={{ background: 'var(--hub-hover)' }}>
                <div
                  className="h-full rounded-full"
                  style={{
                    width: `${(value / maxCat) * 100}%`,
                    background: AMBER,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </motion.div>
  );
};

export default Valuation;
