import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Tag, Trash2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { useBrands, useCreateBrand, useDeleteBrand } from './useBrands';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { toast } from 'sonner';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const AMBER = 'var(--erp-accent)';

const inputStyle: React.CSSProperties = {
  padding: '8px 12px',
  borderRadius: '8px',
  border: '1px solid var(--hub-border)',
  background: 'var(--hub-input-bg)',
  color: 'var(--hub-text)',
  fontSize: '13px',
  outline: 'none',
  width: '100%',
  boxSizing: 'border-box',
};

const Brands: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data: brands = [], isLoading } = useBrands();
  const createMut = useCreateBrand();
  const deleteMut = useDeleteBrand();
  const [showCreate, setShowCreate] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) {
      toast.error(t('erp.brands.nameRequired'));
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        logoUrl: '',
        description: newDescription.trim(),
      });
      toast.success(t('erp.brands.created'));
      setShowCreate(false);
      setNewName('');
      setNewDescription('');
    } catch {
      toast.error(t('erp.brands.createFailed'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/brands'].title))} />
      <PageHeader
        title={t(routeLabels['/erp/brands'].title)}
        description={t('erp.brands.countAcrossCatalog', { count: brands.length })}
        actions={
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: 'var(--dash-text)' }}
          >
            <Plus size={14} /> New Brand
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-24 w-full rounded-xl" />
          ))}
        </div>
      ) : brands.length === 0 ? (
        <EmptyState
          icon={Tag}
          title={t('erp.brands.empty.title')}
          description={t('erp.brands.empty.description')}
          action={{ label: t('erp.brands.newBrand'), onClick: () => setShowCreate(true) }}
        />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {brands.map((b) => (
            <div
              key={b.id}
              className="rounded-xl p-4 transition-colors"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor =
                  'color-mix(in srgb, var(--erp-accent) 30%, transparent)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
              }}
            >
              <div className="flex items-center gap-3 mb-3">
                <div
                  className="h-10 w-10 rounded-lg flex items-center justify-center font-bold text-sm"
                  style={{
                    background: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
                    color: AMBER,
                  }}
                >
                  {(b.name ?? '')
                    .split(' ')
                    .map((w) => w[0])
                    .join('')
                    .slice(0, 2)}
                </div>
                <div className="flex-1 min-w-0">
                  <p
                    className="text-sm font-semibold truncate"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {b.name}
                  </p>
                  <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                    {b.productCount} products
                  </p>
                </div>
                <Tag size={14} style={{ color: 'var(--hub-text-dim)' }} />
                <button
                  type="button"
                  aria-label={t('erp.brands.deleteAria')}
                  onClick={() => setConfirmDeleteId(b.id ?? null)}
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
                  <Trash2 size={13} />
                </button>
              </div>
              {b.description && (
                <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  {b.description}
                </p>
              )}
            </div>
          ))}
        </div>
      )}

      <ConfirmModal
        open={confirmDeleteId !== null}
        title={t('erp.brands.confirmDeleteTitle')}
        description={t('erp.brands.confirmDeleteBody')}
        danger
        confirmLabel={t('erpCommon.delete')}
        onConfirm={async () => {
          if (confirmDeleteId !== null) {
            try {
              await deleteMut.mutateAsync(confirmDeleteId);
              toast.success(t('erp.brands.deleted'));
            } catch {
              // onError in hook handles toast
            }
          }
          setConfirmDeleteId(null);
        }}
        onCancel={() => setConfirmDeleteId(null)}
      />

      {showCreate && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ background: 'var(--hub-overlay)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.brands.newBrand')}
            </h3>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder={t('erp.brands.namePlaceholder')}
              style={inputStyle}
            />
            <input
              value={newDescription}
              onChange={(e) => setNewDescription(e.target.value)}
              placeholder={t('erp.brands.descriptionPlaceholder')}
              style={inputStyle}
            />
            <div className="flex gap-2 justify-end">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                {t('erpCommon.cancel')}
              </button>
              <button
                type="button"
                onClick={handleCreate}
                disabled={createMut.isPending || !newName.trim()}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: 'var(--dash-text)' }}
              >
                {createMut.isPending ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Brands;
