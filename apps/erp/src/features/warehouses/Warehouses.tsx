import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Warehouse as WarehouseIcon, MapPin, User, Trash2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useWarehouses, useCreateWarehouse, useDeleteWarehouse } from './useWarehouses';
import PageHeader from '../../shared/components/PageHeader';
import EmptyState from '../../shared/components/EmptyState';
import { Skeleton } from '../../shared/components/Skeleton';
import ConfirmModal from '../../shared/components/ConfirmModal';
import { formatNumber } from '../../shared/lib/format';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

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
  const { data: warehouses = [], isLoading } = useWarehouses();
  const createMut = useCreateWarehouse();
  const deleteMut = useDeleteWarehouse();
  const [showCreate, setShowCreate] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [newName, setNewName] = useState('');
  const [newAddress, setNewAddress] = useState('');
  const [newManagerName, setNewManagerName] = useState('');
  const [newCapacity, setNewCapacity] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) {
      toast.error('Warehouse name is required');
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        address: newAddress.trim(),
        city: '',
        country: '',
        managerName: newManagerName.trim(),
        capacity: newCapacity.trim() || '0',
        isDefault: false,
        isActive: true,
      });
      toast.success('Warehouse created');
      setShowCreate(false);
      setNewName('');
      setNewAddress('');
      setNewManagerName('');
      setNewCapacity('');
    } catch {
      toast.error('Failed to create warehouse');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Warehouses · rtylr" />
      <PageHeader
        title="Warehouses"
        description={`${warehouses.length} active storage locations`}
        actions={
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Warehouse
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-40 w-full rounded-xl" />
          ))}
        </div>
      ) : warehouses.length === 0 ? (
        <EmptyState
          title="No warehouses"
          description="Create your first warehouse to track stock locations"
        />
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {warehouses.map((w) => {
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
                      style={{ background: 'rgba(245,158,11,0.1)' }}
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
                          style={{ background: 'rgba(16,185,129,0.1)', color: '#10b981' }}
                        >
                          Default
                        </span>
                      )}
                      {!w.isActive && (
                        <span
                          className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                          style={{ background: 'rgba(248,113,113,0.1)', color: '#f87171' }}
                        >
                          Inactive
                        </span>
                      )}
                    </div>
                  </div>
                  <button
                    aria-label="Delete warehouse"
                    onClick={() => setConfirmDeleteId(w.id ?? null)}
                    className="h-7 w-7 rounded-lg flex items-center justify-center"
                    style={{ color: '#f87171' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'rgba(248,113,113,0.1)';
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
                      <User size={11} /> Managed by {w.managerName}
                    </div>
                  )}
                  {Number(w.capacity) > 0 && (
                    <div
                      className="flex items-center justify-between text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <span>Capacity</span>
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
        title="Delete warehouse?"
        description="Products assigned to this warehouse will become unassigned."
        danger
        confirmLabel="Delete"
        onConfirm={async () => {
          if (confirmDeleteId !== null) {
            try {
              await deleteMut.mutateAsync(confirmDeleteId);
              toast.success('Warehouse deleted');
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
          style={{ background: 'rgba(0,0,0,0.5)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              New Warehouse
            </h3>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Warehouse name"
              style={inputStyle}
            />
            <input
              value={newAddress}
              onChange={(e) => setNewAddress(e.target.value)}
              placeholder="Address (optional)"
              style={inputStyle}
            />
            <input
              value={newManagerName}
              onChange={(e) => setNewManagerName(e.target.value)}
              placeholder="Manager name (optional)"
              style={inputStyle}
            />
            <input
              value={newCapacity}
              onChange={(e) => setNewCapacity(e.target.value)}
              placeholder="Capacity (optional)"
              type="number"
              min={0}
              style={inputStyle}
            />
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={createMut.isPending || !newName.trim()}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: '#000' }}
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

export default Warehouses;
