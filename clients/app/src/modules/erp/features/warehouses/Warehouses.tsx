import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Warehouse as WarehouseIcon, MapPin, User, Trash2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useWarehouses, useCreateWarehouse, useDeleteWarehouse } from './useWarehouses';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { formatNumber } from '../../shared/lib/format';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';

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

const Warehouses: React.FC = () => {
  const { data: warehouses, isLoading, isError, refetch } = useWarehouses();
  const warehouseList = warehouses ?? [];
  const createMut = useCreateWarehouse();
  const deleteMut = useDeleteWarehouse();
  const { t } = useTranslation('erp');
  const [showCreate, setShowCreate] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [newName, setNewName] = useState('');
  const [newAddress, setNewAddress] = useState('');
  const [newManagerName, setNewManagerName] = useState('');
  const [newCapacity, setNewCapacity] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) {
      toast.error(t('erp.warehouses.validation.nameRequired'));
      return;
    }
    const capacityText = newCapacity.trim();
    const capacityValue = capacityText === '' ? 0 : Number(capacityText);
    if (!Number.isSafeInteger(capacityValue) || capacityValue < 0) {
      toast.error(t('erpErrors.generic'));
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        address: newAddress.trim(),
        city: '',
        country: '',
        managerName: newManagerName.trim(),
        capacity: String(capacityValue),
        isDefault: false,
        isActive: true,
      });
      toast.success(t('erp.warehouses.toast.created'));
      setShowCreate(false);
      setNewName('');
      setNewAddress('');
      setNewManagerName('');
      setNewCapacity('');
    } catch {
      toast.error(t('erp.warehouses.toast.createFailed'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t('erp.warehouses.title'))} />
      <PageHeader
        title={t('erp.warehouses.title')}
        description={
          isError ? undefined : t('erp.warehouses.count', { count: warehouseList.length })
        }
        actions={
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
          >
            <Plus size={14} /> {t('erp.warehouses.new')}
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-40 w-full rounded-xl" />
          ))}
        </div>
      ) : isError ? (
        <div
          role="alert"
          className="rounded-xl p-5 space-y-3"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <p className="text-sm" style={{ color: 'var(--hub-text)' }}>
            {t('erpErrors.generic')}
          </p>
          <button
            type="button"
            onClick={() => void refetch()}
            className="rounded-lg px-3 py-1.5 text-xs font-semibold"
            style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
          >
            {t('erp.valuation.refresh')}
          </button>
        </div>
      ) : warehouseList.length === 0 ? (
        <EmptyState
          icon={WarehouseIcon}
          title={t('erp.warehouses.empty.title')}
          description={t('erp.warehouses.empty.description')}
          action={{ label: t('erp.warehouses.new'), onClick: () => setShowCreate(true) }}
        />
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {warehouseList.map((w) => {
            const location = [w.city, w.country].filter(Boolean).join(', ');
            return (
              <div
                key={w.id}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div
                      className="h-10 w-10 rounded-lg flex items-center justify-center"
                      style={{
                        background: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)',
                      }}
                    >
                      <WarehouseIcon size={16} style={{ color: AMBER }} />
                    </div>
                    <div className="flex items-center gap-1.5">
                      <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                        {w.name}
                      </p>
                      {w.isDefault && (
                        <span
                          className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                          style={{
                            background: 'color-mix(in srgb, var(--erp-success) 10%, transparent)',
                            color: 'var(--erp-success)',
                          }}
                        >
                          {t('erp.warehouses.status.default')}
                        </span>
                      )}
                      {!w.isActive && (
                        <span
                          className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                          style={{
                            background: 'color-mix(in srgb, var(--erp-danger) 10%, transparent)',
                            color: 'var(--erp-danger)',
                          }}
                        >
                          {t('erp.warehouses.status.inactive')}
                        </span>
                      )}
                    </div>
                  </div>
                  <button
                    type="button"
                    aria-label={t('erp.warehouses.delete')}
                    onClick={() => setConfirmDeleteId(w.id ?? null)}
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
                <div className="space-y-1.5">
                  {(w.address || location) && (
                    <div
                      className="flex items-center gap-2 text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <MapPin size={11} /> {[w.address, location].filter(Boolean).join(' — ')}
                    </div>
                  )}
                  {w.managerName && (
                    <div
                      className="flex items-center gap-2 text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <User size={11} /> {t('erp.warehouses.managedBy', { name: w.managerName })}
                    </div>
                  )}
                  {Number(w.capacity) > 0 && (
                    <div
                      className="flex items-center justify-between text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <span>{t('erp.warehouses.capacity')}</span>
                      <span style={{ color: 'var(--hub-text)' }}>
                        {formatNumber(Number(w.capacity))}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      <ConfirmModal
        open={confirmDeleteId !== null}
        title={t('erp.warehouses.deleteConfirm.title')}
        description={t('erp.warehouses.deleteConfirm.description')}
        danger
        confirmLabel={t('erpCommon.delete')}
        onConfirm={async () => {
          if (confirmDeleteId !== null) {
            try {
              await deleteMut.mutateAsync(confirmDeleteId);
              toast.success(t('erp.warehouses.toast.deleted'));
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
              {t('erp.warehouses.new')}
            </h3>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder={t('erp.warehouses.fields.name')}
              style={inputStyle}
            />
            <input
              value={newAddress}
              onChange={(e) => setNewAddress(e.target.value)}
              placeholder={t('erp.warehouses.fields.address')}
              style={inputStyle}
            />
            <input
              value={newManagerName}
              onChange={(e) => setNewManagerName(e.target.value)}
              placeholder={t('erp.warehouses.fields.manager')}
              style={inputStyle}
            />
            <input
              value={newCapacity}
              onChange={(e) => setNewCapacity(e.target.value)}
              placeholder={t('erp.warehouses.fields.capacity')}
              type="number"
              min={0}
              step={1}
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
                style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
              >
                {createMut.isPending ? t('erp.common.creating') : t('erpCommon.create')}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Warehouses;
