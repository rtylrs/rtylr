import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Tags, Star } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import { useTranslation } from 'react-i18next';
import { usePriceLists, useCreatePriceList } from './usePriceLists';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import { canErpPermission } from '../../shared/lib/erpPermissions';
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

const PriceLists: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data: lists = [], isLoading, refetch } = usePriceLists();
  const createMut = useCreatePriceList();
  const canManagePriceLists = canErpPermission('erp:price_lists:manage');
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newCurrency, setNewCurrency] = useState('USD');
  const [newIsDefault, setNewIsDefault] = useState(false);

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) {
      toast.error(t('erp.priceLists.nameRequired'));
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        currency: newCurrency.trim(),
        isDefault: newIsDefault,
      });
      toast.success(t('erp.priceLists.created'));
      setShowCreate(false);
      setNewName('');
      setNewCurrency('USD');
      setNewIsDefault(false);
    } catch {
      toast.error(t('erp.priceLists.createFailed'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/price-lists'].title))} />
      <PageHeader
        title={t(routeLabels['/erp/price-lists'].title)}
        description={t('erp.priceLists.subtitle')}
        actions={
          canManagePriceLists && (
            <button
              type="button"
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: AMBER, color: 'var(--dash-text)' }}
            >
              <Plus size={14} /> New Price List
            </button>
          )
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-xl" />
          ))}
        </div>
      ) : lists.length === 0 ? (
        <EmptyState
          icon={Tags}
          title={t('erp.priceLists.empty.title')}
          description={t('erp.priceLists.empty.description')}
          action={
            canManagePriceLists
              ? { label: t('erp.priceLists.newPriceList'), onClick: () => setShowCreate(true) }
              : { label: t('erp.priceLists.refresh'), onClick: () => void refetch() }
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {lists.map((l) => (
            <div
              key={l.id}
              className="rounded-xl p-5"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <div
                    className="h-10 w-10 rounded-lg flex items-center justify-center"
                    style={{ background: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)' }}
                  >
                    <Tags size={16} style={{ color: AMBER }} />
                  </div>
                  <div>
                    <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                      {l.name}
                    </p>
                    {l.currency && (
                      <span
                        className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                        style={{
                          background: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
                          color: AMBER,
                        }}
                      >
                        {l.currency}
                      </span>
                    )}
                  </div>
                </div>
                {l.isDefault && (
                  <span
                    className="flex items-center gap-1 text-[10px] font-semibold px-2 py-0.5 rounded-full"
                    style={{
                      background: 'color-mix(in srgb, var(--erp-success) 12%, transparent)',
                      color: 'var(--erp-success)',
                    }}
                  >
                    <Star size={10} /> Default
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

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
              {t('erp.priceLists.newPriceList')}
            </h3>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder={t('erpCommon.name')}
              style={inputStyle}
            />
            <input
              value={newCurrency}
              onChange={(e) => setNewCurrency(e.target.value)}
              placeholder={t('erp.priceLists.currencyPlaceholder')}
              style={inputStyle}
            />
            <label
              className="flex items-center gap-2 text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              <input
                type="checkbox"
                checked={newIsDefault}
                onChange={(e) => setNewIsDefault(e.target.checked)}
              />
              {t('erp.priceLists.setDefault')}
            </label>
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

export default PriceLists;
