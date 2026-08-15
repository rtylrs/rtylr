import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { CheckCircle, Plus, Tag } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import Modal from '../../shared/components/Modal';
import { SkeletonRow } from '../../shared/components/Skeleton';
import {
  useCreateExpenseCategory,
  useExpenseCategories,
} from '../../shared/hooks/useExpenseCategories';
import { cardStyle, ghostBtn, inputStyle, labelStyle, primaryBtn } from '../../shared/lib/tokens';
import { DEFAULT_CATEGORY_COLOR } from '../../shared/lib/settings';

const Categories: React.FC = () => {
  const { data = [], isLoading } = useExpenseCategories();
  const createCategory = useCreateExpenseCategory();
  const [showCreate, setShowCreate] = React.useState(false);
  const [name, setName] = React.useState('');
  const [color, setColor] = React.useState(DEFAULT_CATEGORY_COLOR);

  const handleCreate = (): void => {
    const trimmedName = name.trim();
    if (!trimmedName) return;

    createCategory.mutate(
      { name: trimmedName, color },
      {
        onSuccess: () => {
          setShowCreate(false);
          setName('');
          setColor(DEFAULT_CATEGORY_COLOR);
        },
      }
    );
  };

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
          <button style={primaryBtn} onClick={() => setShowCreate(true)}>
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
            const budgetVal = c.budget ?? 0;
            const spentVal = c.spent ?? 0;
            const pct =
              budgetVal > 0
                ? Math.min(100, Math.round((spentVal / budgetVal) * 100))
                : spentVal > 0
                  ? 100
                  : 0;
            const over = budgetVal > 0 ? spentVal > budgetVal : spentVal > 0;
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

      <Modal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        title="New Expense Category"
        footer={
          <>
            <button
              style={ghostBtn}
              onClick={() => setShowCreate(false)}
              disabled={createCategory.isPending}
            >
              Cancel
            </button>
            <button
              style={{ ...primaryBtn, opacity: createCategory.isPending ? 0.6 : 1 }}
              onClick={handleCreate}
              disabled={createCategory.isPending || !name.trim()}
            >
              <CheckCircle size={13} />
              {createCategory.isPending ? 'Creating…' : 'Create category'}
            </button>
          </>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="expense-category-name">
              Name
            </label>
            <input
              id="expense-category-name"
              autoFocus
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="e.g. Travel"
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="expense-category-color">
              Color
            </label>
            <input
              id="expense-category-color"
              type="color"
              value={color}
              onChange={(event) => setColor(event.target.value)}
              style={{ ...inputStyle, height: 38, padding: 4 }}
            />
          </div>
        </div>
      </Modal>
    </motion.div>
  );
};

export default Categories;
