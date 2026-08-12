import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus, Tag } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useExpenseCategories } from '../../shared/hooks/useExpenseCategories';
import { cardStyle, primaryBtn } from '../../shared/lib/tokens';

const Categories: React.FC = () => {
  const { data = [], isLoading } = useExpenseCategories();

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Expense categories · rtylr" />
      <PageHeader
        title="Expense Categories"
        subtitle={`${data.length} categories`}
        actions={
          <button style={primaryBtn} disabled title="Category management coming in a future update">
            <Plus size={13} /> New Category
          </button>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState
          icon={Tag}
          title="No categories yet"
          description="Create categories to organize expenses and set budgets."
        />
      ) : (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: 14,
          }}
        >
          {data.map((c) => {
            const pct = Math.min(100, Math.round(((c.spent ?? 0) / (c.budget ?? 1)) * 100));
            const over = (c.spent ?? 0) > (c.budget ?? 0);
            return (
              <div key={c.id} style={{ ...cardStyle, padding: 16 }}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: 10,
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <span style={{ width: 10, height: 10, borderRadius: 5, background: c.color }} />
                    <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
                      {c.name}
                    </span>
                  </div>
                  <span
                    style={{
                      fontSize: 11,
                      color: over ? 'var(--finance-error)' : 'var(--hub-text-dim)',
                    }}
                  >
                    {pct}%
                  </span>
                </div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    fontSize: 12,
                    color: 'var(--hub-text-muted)',
                    marginBottom: 8,
                  }}
                >
                  <span>
                    <MoneyDisplay amount={c.spent} /> of <MoneyDisplay amount={c.budget} />
                  </span>
                </div>
                <div
                  style={{
                    height: 6,
                    background: 'var(--hub-hover)',
                    borderRadius: 3,
                    overflow: 'hidden',
                  }}
                >
                  <div
                    style={{
                      width: `${pct}%`,
                      height: '100%',
                      background: over ? 'var(--finance-error)' : c.color,
                      transition: 'width 0.3s',
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default Categories;
