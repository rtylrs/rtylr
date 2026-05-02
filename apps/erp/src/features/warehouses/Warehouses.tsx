import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Warehouse as WarehouseIcon, MapPin, User, Trash2 } from 'lucide-react';
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
  const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);
  const [newName, setNewName] = useState('');
  const [newAddress, setNewAddress] = useState('');
  const [newManager, setNewManager] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) return;
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        address: newAddress.trim(),
        manager: newManager.trim(),
      });
      toast.success('Warehouse created');
      setShowCreate(false);
      setNewName('');
      setNewAddress('');
      setNewManager('');
    } catch {
      toast.error('Failed to create warehouse');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
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
            const pct = Math.round((w.used / w.capacity) * 100);
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
                    <div>
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
                    </div>
                  </div>
                  <button
                    aria-label="Delete warehouse"
                    onClick={() => setConfirmDeleteId(w.id)}
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
                <div className="space-y-1.5 mb-4">
                  <div
                    className="flex items-center gap-2 text-xs"
                    style={{ color: 'var(--hub-text-muted)' }}
                  >
                    <MapPin size={11} /> {w.address}
                  </div>
                  <div
                    className="flex items-center gap-2 text-xs"
                    style={{ color: 'var(--hub-text-muted)' }}
                  >
                    <User size={11} /> Managed by {w.manager}
                  </div>
                </div>
                <div>
                  <div className="flex justify-between text-xs mb-1">
                    <span style={{ color: 'var(--hub-text-muted)' }}>Capacity</span>
                    <span style={{ color: 'var(--hub-text)' }}>
                      {formatNumber(w.used)} / {formatNumber(w.capacity)}
                    </span>
                  </div>
                  <div
                    className="h-2 rounded-full overflow-hidden"
                    style={{ background: 'var(--hub-hover)' }}
                  >
                    <div
                      className="h-full rounded-full"
                      style={{
                        width: `${pct}%`,
                        background: pct > 85 ? '#f87171' : pct > 65 ? AMBER : '#10b981',
                      }}
                    />
                  </div>
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
            await deleteMut.mutateAsync(confirmDeleteId);
            toast.success('Warehouse deleted');
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
              value={newManager}
              onChange={(e) => setNewManager(e.target.value)}
              placeholder="Manager name (optional)"
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
