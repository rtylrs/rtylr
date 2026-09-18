import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { CheckCircle, Edit2, Plus, Tag, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import Modal from '../../shared/components/Modal';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { SkeletonRow } from '../../shared/components/Skeleton';
import {
  useCreateExpenseCategory,
  useDeleteExpenseCategory,
  useExpenseCategories,
  useUpdateExpenseCategory,
} from '../../shared/hooks/useExpenseCategories';
import { cardStyle, ghostBtn, inputStyle, labelStyle, primaryBtn } from '../../shared/lib/tokens';
import { DEFAULT_CATEGORY_COLOR } from '../../shared/lib/settings';
import type { ExpenseCategory } from '../../shared/services/financeApi';

const Categories: React.FC = () => {
  const { t } = useTranslation();
  const { data = [], isLoading } = useExpenseCategories();
  const createCategory = useCreateExpenseCategory();
  const updateCategory = useUpdateExpenseCategory();
  const deleteCategory = useDeleteExpenseCategory();

  const [showCreate, setShowCreate] = useState(false);
  const [editingCategory, setEditingCategory] = useState<ExpenseCategory | null>(null);
  const [deletingCategoryId, setDeletingCategoryId] = useState<number | null>(null);

  const [name, setName] = useState('');
  const [color, setColor] = useState(DEFAULT_CATEGORY_COLOR);
  const [budget, setBudget] = useState<string>('');

  const [editName, setEditName] = useState('');
  const [editColor, setEditColor] = useState(DEFAULT_CATEGORY_COLOR);
  const [editBudget, setEditBudget] = useState<string>('');

  const handleCreate = (): void => {
    const trimmedName = name.trim();
    if (!trimmedName) return;

    createCategory.mutate(
      {
        name: trimmedName,
        color,
        budget: budget ? parseFloat(budget) : undefined,
      },
      {
        onSuccess: () => {
          setShowCreate(false);
          setName('');
          setColor(DEFAULT_CATEGORY_COLOR);
          setBudget('');
        },
      }
    );
  };

  const startEdit = (c: ExpenseCategory): void => {
    setEditingCategory(c);
    setEditName(c.name);
    setEditColor(c.color || DEFAULT_CATEGORY_COLOR);
    setEditBudget(c.budget ? String(c.budget) : '');
  };

  const handleUpdate = (): void => {
    if (!editingCategory) return;
    const trimmed = editName.trim();
    if (!trimmed) return;

    updateCategory.mutate(
      {
        id: editingCategory.id,
        data: {
          name: trimmed,
          color: editColor,
          budget: editBudget ? parseFloat(editBudget) : 0,
        },
      },
      {
        onSuccess: () => {
          setEditingCategory(null);
        },
      }
    );
  };

  const handleDelete = (): void => {
    if (deletingCategoryId == null) return;
    deleteCategory.mutate(deletingCategoryId, {
      onSuccess: () => {
        setDeletingCategoryId(null);
      },
    });
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('common.pageTitles.expenseCategories'))} />
      <PageHeader
        title={t('categories.title')}
        subtitle={`${data.length} categories`}
        actions={
          <button type="button" style={primaryBtn} onClick={() => setShowCreate(true)}>
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
          title={t('categories.emptyTitle')}
          description={t('categories.emptyDescription')}
          action={{
            label: t('categories.emptyCta'),
            onClick: () => setShowCreate(true),
          }}
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
                  <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    <button
                      type="button"
                      onClick={() => startEdit(c)}
                      className="p-1 rounded hover:bg-white/10 text-white/60 hover:text-white transition"
                      title={t('categories.edit')}
                      aria-label={`Edit ${c.name}`}
                    >
                      <Edit2 size={13} />
                    </button>
                    <button
                      type="button"
                      onClick={() => setDeletingCategoryId(c.id)}
                      className="p-1 rounded hover:bg-red-500/10 text-white/60 hover:text-red-400 transition"
                      title={t('categories.delete')}
                      aria-label={`Delete ${c.name}`}
                    >
                      <Trash2 size={13} />
                    </button>
                    <span
                      style={{
                        fontSize: 11,
                        color: over ? 'var(--finance-error)' : 'var(--hub-text-dim)',
                        marginLeft: 4,
                      }}
                    >
                      {pct}%
                    </span>
                  </div>
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

      {/* Create Modal */}
      <Modal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        title={t('categories.newTitle')}
        footer={
          <>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setShowCreate(false)}
              disabled={createCategory.isPending}
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
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
              {t('categories.name')}
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
            <label style={labelStyle} htmlFor="expense-category-budget">
              Monthly Budget (Optional)
            </label>
            <input
              id="expense-category-budget"
              type="number"
              min="0"
              step="1"
              value={budget}
              onChange={(event) => setBudget(event.target.value)}
              placeholder="e.g. 500"
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="expense-category-color">
              {t('categories.color')}
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

      {/* Edit Modal */}
      <Modal
        open={Boolean(editingCategory)}
        onClose={() => setEditingCategory(null)}
        title={t('categories.editTitle')}
        footer={
          <>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setEditingCategory(null)}
              disabled={updateCategory.isPending}
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
              style={{ ...primaryBtn, opacity: updateCategory.isPending ? 0.6 : 1 }}
              onClick={handleUpdate}
              disabled={updateCategory.isPending || !editName.trim()}
            >
              <CheckCircle size={13} />
              {updateCategory.isPending ? 'Saving…' : 'Save Changes'}
            </button>
          </>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="edit-category-name">
              {t('categories.name')}
            </label>
            <input
              id="edit-category-name"
              autoFocus
              value={editName}
              onChange={(event) => setEditName(event.target.value)}
              placeholder="e.g. Travel"
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="edit-category-budget">
              {t('categories.monthlyBudget')}
            </label>
            <input
              id="edit-category-budget"
              type="number"
              min="0"
              step="1"
              value={editBudget}
              onChange={(event) => setEditBudget(event.target.value)}
              placeholder="e.g. 500"
              style={inputStyle}
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle} htmlFor="edit-category-color">
              {t('categories.color')}
            </label>
            <input
              id="edit-category-color"
              type="color"
              value={editColor}
              onChange={(event) => setEditColor(event.target.value)}
              style={{ ...inputStyle, height: 38, padding: 4 }}
            />
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={deletingCategoryId != null}
        title={t('categories.deleteTitle')}
        message={t('categories.deleteConfirm')}
        confirmLabel={t('common.delete')}
        destructive
        onConfirm={handleDelete}
        onClose={() => setDeletingCategoryId(null)}
      />
    </motion.div>
  );
};

export default Categories;
