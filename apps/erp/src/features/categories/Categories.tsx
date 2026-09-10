import React, { useRef, useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Folder, Edit3, Trash2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import {
  useCategories,
  useCreateCategory,
  useDeleteCategory,
  useUpdateCategory,
} from './useCategories';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { toast } from 'sonner';

const AMBER = 'var(--erp-accent)';

const Categories: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data: categories = [], isLoading } = useCategories();
  const createMut = useCreateCategory();
  const deleteMut = useDeleteCategory();
  const updateMut = useUpdateCategory();
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const newNameInputRef = useRef<HTMLInputElement>(null);

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) {
      toast.error('Category name is required');
      return;
    }
    try {
      await createMut.mutateAsync({ name: newName.trim(), description: '' });
      toast.success(`Category "${newName}" created`);
      setNewName('');
    } catch {
      toast.error('Failed to create category');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Categories · rtylr" />
      <PageHeader title="Categories" description="Organize products into hierarchical categories" />

      <div
        className="rounded-xl p-4 flex items-center gap-2"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <input
          ref={newNameInputRef}
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
          placeholder="New category name…"
          className="flex-1 px-3 py-2 rounded-lg text-sm outline-none"
          style={{
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text)',
          }}
        />
        <button
          type="button"
          onClick={handleCreate}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
        >
          <Plus size={14} /> Add
        </button>
      </div>

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        {isLoading ? (
          <div className="p-4 space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : categories.length === 0 ? (
          <EmptyState
            icon={Folder}
            title={t('erp.categories.empty.title')}
            description={t('erp.categories.empty.description')}
            action={{
              label: t('erp.categories.addCategory'),
              onClick: () => newNameInputRef.current?.focus(),
            }}
          />
        ) : (
          categories.map((cat, i) => (
            <div
              key={cat.id}
              className="flex items-center gap-3 px-4 py-3"
              style={{
                borderBottom: i < categories.length - 1 ? '1px solid var(--hub-border)' : 'none',
              }}
            >
              <div
                className="flex items-center justify-center h-8 w-8 rounded-lg"
                style={{ background: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)' }}
              >
                <Folder size={14} style={{ color: AMBER }} />
              </div>
              <div className="flex-1">
                {editingId === cat.id ? (
                  <input
                    value={editName}
                    onChange={(e) => setEditName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        updateMut.mutate(
                          {
                            id: cat.id,
                            body: { name: editName.trim(), description: cat.description },
                          },
                          {
                            onSuccess: () => {
                              toast.success('Category updated');
                              setEditingId(null);
                            },
                          }
                        );
                      }
                      if (e.key === 'Escape') setEditingId(null);
                    }}
                    className="flex-1 px-3 py-1 rounded-lg text-sm outline-none"
                    style={{
                      background: 'var(--hub-input-bg)',
                      border: '1px solid var(--hub-border)',
                      color: 'var(--hub-text)',
                    }}
                    autoFocus
                  />
                ) : (
                  <>
                    <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                      {cat.name}
                    </p>
                    {cat.description && (
                      <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                        {cat.description}
                      </p>
                    )}
                  </>
                )}
              </div>
              {editingId === cat.id ? (
                <>
                  <button
                    type="button"
                    aria-label="Save category"
                    onClick={() =>
                      updateMut.mutate(
                        {
                          id: cat.id,
                          body: { name: editName.trim(), description: cat.description },
                        },
                        {
                          onSuccess: () => {
                            toast.success('Category updated');
                            setEditingId(null);
                          },
                        }
                      )
                    }
                    className="h-7 w-7 rounded-lg flex items-center justify-center text-xs font-bold"
                    style={{ color: 'var(--erp-success)' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background =
                        'color-mix(in srgb, var(--erp-success) 10%, transparent)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    ✓
                  </button>
                  <button
                    type="button"
                    aria-label="Cancel edit"
                    onClick={() => setEditingId(null)}
                    className="h-7 w-7 rounded-lg flex items-center justify-center text-xs font-bold"
                    style={{ color: 'var(--erp-danger)' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background =
                        'color-mix(in srgb, var(--erp-danger) 10%, transparent)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    ✗
                  </button>
                </>
              ) : (
                <>
                  <button
                    type="button"
                    aria-label="Edit category"
                    onClick={() => {
                      setEditingId(cat.id ?? null);
                      setEditName(cat.name ?? '');
                    }}
                    className="h-7 w-7 rounded-lg flex items-center justify-center"
                    style={{ color: 'var(--hub-text-muted)' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <Edit3 size={12} />
                  </button>
                  <button
                    type="button"
                    onClick={() => setConfirmDelete(cat.id ?? null)}
                    className="h-7 w-7 rounded-lg flex items-center justify-center"
                    style={{ color: 'var(--erp-danger)' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background =
                        'color-mix(in srgb, var(--erp-danger) 10%, transparent)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <Trash2 size={12} />
                  </button>
                </>
              )}
            </div>
          ))
        )}
      </div>

      <ConfirmModal
        open={confirmDelete !== null}
        title="Delete category?"
        description="Products in this category will not be deleted, but they will become uncategorized."
        danger
        confirmLabel="Delete"
        onConfirm={async () => {
          if (confirmDelete !== null) {
            try {
              await deleteMut.mutateAsync(confirmDelete);
              toast.success('Category deleted');
            } catch {
              // onError in hook handles toast
            }
          }
          setConfirmDelete(null);
        }}
        onCancel={() => setConfirmDelete(null)}
      />
    </motion.div>
  );
};

export default Categories;
