import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { CheckCircle, Edit2, Plus, Tag, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import PageHeader from '../../shared/components/PageHeader';
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
  const { t } = useTranslation('financeCommon');
  const { data = [], isLoading } = useExpenseCategories();
  const createCategory = useCreateExpenseCategory();
  const updateCategory = useUpdateExpenseCategory();
  const deleteCategory = useDeleteExpenseCategory();

  const [showCreate, setShowCreate] = useState(false);
  const [editingCategory, setEditingCategory] = useState<ExpenseCategory | null>(null);
  const [deletingCategoryId, setDeletingCategoryId] = useState<number | null>(null);

  const [name, setName] = useState('');
  const [color, setColor] = useState(DEFAULT_CATEGORY_COLOR);

  const [editName, setEditName] = useState('');
  const [editColor, setEditColor] = useState(DEFAULT_CATEGORY_COLOR);

  const handleCreate = (): void => {
    const trimmedName = name.trim();
    if (!trimmedName) return;

    createCategory.mutate(
      {
        name: trimmedName,
        color,
      },
      {
        onSuccess: () => {
          setShowCreate(false);
          setName('');
          setColor(DEFAULT_CATEGORY_COLOR);
        },
      }
    );
  };

  const startEdit = (c: ExpenseCategory): void => {
    setEditingCategory(c);
    setEditName(c.name);
    setEditColor(c.color || DEFAULT_CATEGORY_COLOR);
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
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.expenseCategories'))} />
      <PageHeader
        title={t('financeCategories.title')}
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
          title={t('financeCategories.emptyTitle')}
          description={t('financeCategories.emptyDescription')}
          action={{
            label: t('financeCategories.emptyCta'),
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
          {data.map((c) => (
            <div key={c.id} style={{ ...cardStyle, padding: 16 }}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
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
                    title={t('financeCategories.edit')}
                    aria-label={`Edit ${c.name}`}
                  >
                    <Edit2 size={13} />
                  </button>
                  <button
                    type="button"
                    onClick={() => setDeletingCategoryId(c.id)}
                    className="p-1 rounded hover:bg-red-500/10 text-white/60 hover:text-red-400 transition"
                    title={t('financeCategories.delete')}
                    aria-label={`Delete ${c.name}`}
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      <Modal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        title={t('financeCategories.newTitle')}
        footer={
          <>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setShowCreate(false)}
              disabled={createCategory.isPending}
            >
              {t('financeCommon.cancel')}
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
              {t('financeCategories.name')}
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
              {t('financeCategories.color')}
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
        title={t('financeCategories.editTitle')}
        footer={
          <>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setEditingCategory(null)}
              disabled={updateCategory.isPending}
            >
              {t('financeCommon.cancel')}
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
              {t('financeCategories.name')}
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
            <label style={labelStyle} htmlFor="edit-category-color">
              {t('financeCategories.color')}
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
        title={t('financeCategories.deleteTitle')}
        message={t('financeCategories.deleteConfirm')}
        confirmLabel={t('financeCommon.delete')}
        destructive
        onConfirm={handleDelete}
        onClose={() => setDeletingCategoryId(null)}
      />
    </motion.div>
  );
};

export default Categories;
